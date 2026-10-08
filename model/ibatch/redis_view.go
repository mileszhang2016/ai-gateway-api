// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License. All rights reserved.

package ibatch

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// redisTaskView 是 Redis BATCH_TASK hash 的读视图。指针字段区分
// "hash 未携带"（nil，不覆盖 DB 已有值）与零值。
//
// 键结构（批量任务与对账.md §2）：{api_key_id, provider, key_name,
// input_file_id, est_lines, reserved_units, status, usage_in, usage_out,
// settle_units, settle_status}。数据面一期实现以 cluster 字段承载路由
// cluster 名（provider 列同值），故 provider 缺失时回退读 cluster；
// 预留布尔标记为 reserved（"1"/"0"），reserved_units 金额优先。
type redisTaskView struct {
	batchID      string
	apiKeyID     string
	provider     string
	keyName      string
	inputFileID  string
	estLines     *int64
	reserveUnits *int64
	status       string
	usageIn      *int64
	usageOut     *int64
	settleUnits  *int64
	settleStatus string
}

// ParseRedisTaskView 解析 BATCH_TASK hash；batch_id 取自键。
// status 缺失视为脏数据返回错误（调用方跳过并告警）。
func ParseRedisTaskView(batchID string, vals map[string]string) (*redisTaskView, error) {
	if batchID == "" {
		return nil, fmt.Errorf("empty batch id")
	}
	status := vals["status"]
	if status == "" {
		return nil, fmt.Errorf("BATCH_TASK:%s missing status", batchID)
	}
	view := &redisTaskView{
		batchID:      batchID,
		apiKeyID:     vals["api_key_id"],
		keyName:      vals["key_name"],
		inputFileID:  vals["input_file_id"],
		status:       status,
		settleStatus: vals["settle_status"],
	}
	view.provider = vals["provider"]
	if view.provider == "" {
		// 兼容数据面一期字段名（路由 cluster 名，与 provider 列同值）。
		view.provider = vals["cluster"]
	}
	if v, ok := parseInt64Field(vals, "est_lines"); ok {
		view.estLines = &v
	}
	if v, ok := parseInt64Field(vals, "usage_in"); ok {
		view.usageIn = &v
	}
	if v, ok := parseInt64Field(vals, "usage_out"); ok {
		view.usageOut = &v
	}
	if v, ok := parseInt64Field(vals, "settle_units"); ok {
		view.settleUnits = &v
	}
	if v, ok := parseInt64Field(vals, "reserved_units"); ok {
		view.reserveUnits = &v
	} else if vals["reserved"] == "1" {
		// 数据面一期仅写 reserved 布尔标记（金额簿记在
		// BATCH_RESERVE_BATCH），预留金额保持未知（0）。
		zero := int64(0)
		view.reserveUnits = &zero
	}
	return view, nil
}

// redisFileView 是 Redis BATCH_FILE hash 的读视图。键结构：
// {api_key_id, key_name, lines, bytes, purpose, dir}（输出文件绑定另含
// batch_id）。
type redisFileView struct {
	partition string
	fileID    string
	apiKeyID  string
	keyName   string
	lines     *int64
	bytes     *int64
	purpose   string
	direction string
}

// ParseRedisFileView 解析 BATCH_FILE hash。
func ParseRedisFileView(partition, fileID string, vals map[string]string) *redisFileView {
	view := &redisFileView{
		partition: partition,
		fileID:    fileID,
		apiKeyID:  vals["api_key_id"],
		keyName:   vals["key_name"],
		purpose:   vals["purpose"],
		direction: vals["dir"],
	}
	if v, ok := parseInt64Field(vals, "lines"); ok {
		view.lines = &v
	}
	if v, ok := parseInt64Field(vals, "bytes"); ok {
		view.bytes = &v
	}
	return view
}

func parseInt64Field(vals map[string]string, field string) (int64, bool) {
	raw, ok := vals[field]
	if !ok || raw == "" {
		return 0, false
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// applyTaskView 把 Redis 视图组装为任务实体：经 api_key 解析
// product_name / entity_id。api_key 不存在返回 (nil, nil)（脏数据跳过，
// 由对账巡检日志补登兜底）。
func (m *Manager) applyTaskView(ctx context.Context, cache *apiKeyResolveCache, view *redisTaskView) (*BatchTask, error) {
	ak, err := cache.resolve(ctx, m.apiKeyQuerier, view.apiKeyID)
	if err != nil {
		return nil, err
	}
	if ak == nil {
		return nil, nil
	}
	task := &BatchTask{
		BatchID:     view.batchID,
		IdemKey:     view.batchID,
		APIKeyID:    view.apiKeyID,
		Provider:    view.provider,
		KeyName:     view.keyName,
		InputFileID: view.inputFileID,
		Status:      view.status,
		SyncSource:  SyncSourceRedis,
	}
	if ak.ProductName != nil {
		task.ProductName = *ak.ProductName
	}
	if ak.EntityID != nil {
		task.EntityID = *ak.EntityID
	}
	if view.estLines != nil {
		task.EstLines = *view.estLines
	}
	if view.reserveUnits != nil {
		task.ReserveUnits = *view.reserveUnits
	}
	if view.usageIn != nil {
		task.UsageInputTokens = *view.usageIn
	}
	if view.usageOut != nil {
		task.UsageOutputTokens = *view.usageOut
	}
	if view.settleUnits != nil {
		task.SettleUnits = *view.settleUnits
	}
	if view.settleStatus != "" {
		task.SettleStatus = view.settleStatus
	} else {
		task.SettleStatus = SettleStatusReserved
	}
	if IsTerminalStatus(task.Status) {
		now := m.clock.Now()
		task.TerminalAt = &now
	}
	return task, nil
}

// applyFileView 把 Redis 视图组装为文件实体（同 applyTaskView 的
// product/entity 解析语义）。
func (m *Manager) applyFileView(ctx context.Context, cache *apiKeyResolveCache, view *redisFileView) (*BatchFile, error) {
	ak, err := cache.resolve(ctx, m.apiKeyQuerier, view.apiKeyID)
	if err != nil {
		return nil, err
	}
	if ak == nil {
		return nil, nil
	}
	file := &BatchFile{
		FileID:    view.fileID,
		Provider:  view.partition,
		APIKeyID:  view.apiKeyID,
		KeyName:   view.keyName,
		Direction: view.direction,
		Purpose:   view.purpose,
	}
	if ak.ProductName != nil {
		file.ProductName = *ak.ProductName
	}
	if view.lines != nil {
		file.Lines = *view.lines
	}
	if view.bytes != nil {
		file.Bytes = *view.bytes
	}
	if file.Direction == "" {
		file.Direction = FileDirectionInput
	}
	return file, nil
}

// mergeTaskForSync 合并 Redis 视图与 DB 已有行（终态不回退；hash 未
// 携带的字段不覆盖 DB 已有值）。existing 为 nil 时按 insert 语义组装
// （terminal 状态记 terminal_at = now 近似）。
//
// 规则：
//   - existing 已终态：status / terminal_at 保持 DB 值，其余可变字段
//     照常刷新（BFE 原地更新的 usage/settle 仍要取用）；
//   - 非终态 → 终态：记 terminal_at = now（同步时刻近似进入终态时间）；
//   - 数值字段仅在 view 携带（hash 存在该字段）时刷新，避免同步把
//     DB 已有值抹零。
func mergeTaskForSync(existing *BatchTask, incoming *BatchTask, view *redisTaskView, now time.Time) *BatchTask {
	if existing == nil {
		if IsTerminalStatus(incoming.Status) {
			incoming.TerminalAt = &now
		}
		return incoming
	}
	merged := *incoming
	merged.ID = existing.ID
	merged.IdemKey = existing.IdemKey
	merged.BatchID = existing.BatchID
	// 归属与 insert 时事实不回退（api_key 删除后 Redis 同步不得抹掉归属）。
	if merged.APIKeyID == "" {
		merged.APIKeyID = existing.APIKeyID
	}
	if merged.ProductName == "" {
		merged.ProductName = existing.ProductName
	}
	if merged.EntityID == "" {
		merged.EntityID = existing.EntityID
	}
	if merged.Provider == "" {
		merged.Provider = existing.Provider
	}
	if merged.KeyName == "" {
		merged.KeyName = existing.KeyName
	}
	if merged.InputFileID == "" {
		merged.InputFileID = existing.InputFileID
	}
	if merged.OutputFileID == "" {
		merged.OutputFileID = existing.OutputFileID
	}
	if merged.Endpoint == "" {
		merged.Endpoint = existing.Endpoint
	}
	if merged.UsageSource == "" {
		merged.UsageSource = existing.UsageSource
	}
	if merged.RequestCounts == nil {
		merged.RequestCounts = existing.RequestCounts
	}
	// sync_source 不回退：log 补登标记保留供巡检告警。
	if existing.SyncSource != "" {
		merged.SyncSource = existing.SyncSource
	}
	// 数值字段按 hash 存在性刷新。
	if view.estLines != nil {
		merged.EstLines = *view.estLines
	} else {
		merged.EstLines = existing.EstLines
	}
	if view.reserveUnits != nil {
		merged.ReserveUnits = *view.reserveUnits
	} else {
		merged.ReserveUnits = existing.ReserveUnits
	}
	if view.usageIn != nil {
		merged.UsageInputTokens = *view.usageIn
	} else {
		merged.UsageInputTokens = existing.UsageInputTokens
	}
	if view.usageOut != nil {
		merged.UsageOutputTokens = *view.usageOut
	} else {
		merged.UsageOutputTokens = existing.UsageOutputTokens
	}
	if view.settleUnits != nil {
		merged.SettleUnits = *view.settleUnits
	} else {
		merged.SettleUnits = existing.SettleUnits
	}
	if view.settleStatus == "" {
		merged.SettleStatus = existing.SettleStatus
	}
	// 终态不回退。
	if IsTerminalStatus(existing.Status) {
		merged.Status = existing.Status
		merged.TerminalAt = existing.TerminalAt
		return &merged
	}
	if IsTerminalStatus(merged.Status) {
		merged.TerminalAt = &now
	}
	return &merged
}

// mergeFileForSync 合并 Redis 文件视图与 DB 已有行（幂等刷新
// last_seen_at；归属列不回退；lines/bytes 按 hash 存在性刷新）。
func mergeFileForSync(existing *BatchFile, incoming *BatchFile, view *redisFileView, now time.Time) *BatchFile {
	if existing == nil {
		if incoming.LastSeenAt == nil {
			incoming.LastSeenAt = &now
		}
		if incoming.FirstSeenAt == nil {
			incoming.FirstSeenAt = &now
		}
		return incoming
	}
	merged := *incoming
	merged.ID = existing.ID
	if merged.APIKeyID == "" {
		merged.APIKeyID = existing.APIKeyID
	}
	if merged.ProductName == "" {
		merged.ProductName = existing.ProductName
	}
	if merged.KeyName == "" {
		merged.KeyName = existing.KeyName
	}
	if merged.Direction == "" {
		merged.Direction = existing.Direction
	}
	if merged.Purpose == "" {
		merged.Purpose = existing.Purpose
	}
	if view.lines != nil {
		merged.Lines = *view.lines
	} else {
		merged.Lines = existing.Lines
	}
	if view.bytes != nil {
		merged.Bytes = *view.bytes
	} else {
		merged.Bytes = existing.Bytes
	}
	merged.FirstSeenAt = existing.FirstSeenAt
	merged.LastSeenAt = &now
	return &merged
}
