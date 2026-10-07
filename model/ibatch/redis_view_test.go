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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBatchTaskKey(t *testing.T) {
	id, ok := ParseBatchTaskKey("BATCH_TASK:batch_abc")
	require.True(t, ok)
	assert.Equal(t, "batch_abc", id)

	_, ok = ParseBatchTaskKey("BATCH_FILE:x:y")
	assert.False(t, ok)
	_, ok = ParseBatchTaskKey("BATCH_TASK:")
	assert.False(t, ok)
}

func TestParseBatchFileKey(t *testing.T) {
	partition, fileID, ok := ParseBatchFileKey("BATCH_FILE:cluster-a:file-1")
	require.True(t, ok)
	assert.Equal(t, "cluster-a", partition)
	assert.Equal(t, "file-1", fileID)

	_, _, ok = ParseBatchFileKey("BATCH_FILE:nocolon")
	assert.False(t, ok)
	_, _, ok = ParseBatchFileKey("OTHER:prefix")
	assert.False(t, ok)
}

func TestRedisKeyBuilders(t *testing.T) {
	assert.Equal(t, "BATCH_TASK:b1", BatchTaskKey("b1"))
	assert.Equal(t, "BATCH_FILE:p:f1", BatchFileKey("p", "f1"))
	assert.Equal(t, "BATCH_SETTLED:b1", BatchSettledKey("b1"))
	assert.Equal(t, "BATCH_RESERVE_BATCH:b1", BatchReserveBatchKey("b1"))
	assert.Equal(t, "BATCH_RESERVE:plan-x", BatchReserveMirrorKey("plan-x"))
	assert.Equal(t, "BATCH_RESERVE_BATCH:*", ReserveBatchKeyPattern())
}

func TestParseRedisTaskView(t *testing.T) {
	t.Run("全字段", func(t *testing.T) {
		view, err := ParseRedisTaskView("b-1", map[string]string{
			"api_key_id":     "ak",
			"provider":       "openai",
			"key_name":       "k1",
			"input_file_id":  "f-in",
			"est_lines":      "10",
			"reserved_units": "500",
			"status":         "in_progress",
			"usage_in":       "1",
			"usage_out":      "2",
			"settle_units":   "0",
			"settle_status":  "reserved",
		})
		require.NoError(t, err)
		assert.Equal(t, "openai", view.provider)
		require.NotNil(t, view.reserveUnits)
		assert.Equal(t, int64(500), *view.reserveUnits)
		require.NotNil(t, view.usageIn)
		assert.Equal(t, int64(1), *view.usageIn)
	})

	t.Run("provider 缺失回退 cluster（数据面一期字段名）", func(t *testing.T) {
		view, err := ParseRedisTaskView("b-1", map[string]string{
			"api_key_id": "ak", "cluster": "c-1", "status": "queued",
		})
		require.NoError(t, err)
		assert.Equal(t, "c-1", view.provider)
	})

	t.Run("reserved 布尔标记兜底（无金额字段）", func(t *testing.T) {
		view, err := ParseRedisTaskView("b-1", map[string]string{
			"api_key_id": "ak", "status": "queued", "reserved": "1",
		})
		require.NoError(t, err)
		require.NotNil(t, view.reserveUnits)
		assert.Equal(t, int64(0), *view.reserveUnits)
	})

	t.Run("status 缺失报脏数据", func(t *testing.T) {
		_, err := ParseRedisTaskView("b-1", map[string]string{"api_key_id": "ak"})
		require.Error(t, err)
	})

	t.Run("非法数值字段跳过", func(t *testing.T) {
		view, err := ParseRedisTaskView("b-1", map[string]string{
			"api_key_id": "ak", "status": "queued", "usage_in": "NaN",
		})
		require.NoError(t, err)
		assert.Nil(t, view.usageIn)
	})
}

func TestParseRedisFileView(t *testing.T) {
	view := ParseRedisFileView("p", "f-1", map[string]string{
		"api_key_id": "ak",
		"key_name":   "k1",
		"lines":      "8",
		"bytes":      "999",
		"purpose":    "batch",
		"dir":        "output",
	})
	assert.Equal(t, "p", view.partition)
	assert.Equal(t, "f-1", view.fileID)
	require.NotNil(t, view.lines)
	assert.Equal(t, int64(8), *view.lines)
	assert.Equal(t, "output", view.direction)
}

func TestMergeTaskForSync(t *testing.T) {
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)

	newIncoming := func() *BatchTask {
		return &BatchTask{
			BatchID: "b-1", IdemKey: "b-1", APIKeyID: "ak", ProductName: "p1",
			Provider: "openai", Status: TaskStatusInProgress, SettleStatus: SettleStatusReserved,
			SyncSource: SyncSourceRedis,
		}
	}

	t.Run("existing 为 nil：终态记 terminal_at", func(t *testing.T) {
		incoming := newIncoming()
		incoming.Status = TaskStatusCompleted
		merged := mergeTaskForSync(nil, incoming, &redisTaskView{batchID: "b-1", status: "completed"}, now)
		assert.Equal(t, TaskStatusCompleted, merged.Status)
		require.NotNil(t, merged.TerminalAt)
		assert.Equal(t, now, *merged.TerminalAt)
	})

	t.Run("终态不回退", func(t *testing.T) {
		terminalAt := now.Add(-time.Hour)
		existing := newIncoming()
		existing.ID = 5
		existing.Status = TaskStatusCompleted
		existing.TerminalAt = &terminalAt
		existing.SyncSource = SyncSourceLog

		incoming := newIncoming()
		incoming.Status = TaskStatusInProgress // Redis 滞后状态
		view := &redisTaskView{batchID: "b-1", status: "in_progress"}

		merged := mergeTaskForSync(existing, incoming, view, now)
		assert.Equal(t, TaskStatusCompleted, merged.Status)
		assert.Equal(t, terminalAt, *merged.TerminalAt)
		assert.Equal(t, int64(5), merged.ID)
		assert.Equal(t, SyncSourceLog, merged.SyncSource, "log 补登标记不得回退")
	})

	t.Run("非终态到终态记 terminal_at=now", func(t *testing.T) {
		existing := newIncoming()
		existing.ID = 6
		incoming := newIncoming()
		incoming.Status = TaskStatusCompleted
		view := &redisTaskView{batchID: "b-1", status: "completed"}
		merged := mergeTaskForSync(existing, incoming, view, now)
		assert.Equal(t, TaskStatusCompleted, merged.Status)
		require.NotNil(t, merged.TerminalAt)
		assert.Equal(t, now, *merged.TerminalAt)
	})

	t.Run("数值字段按 hash 存在性刷新", func(t *testing.T) {
		existing := newIncoming()
		existing.ID = 7
		existing.UsageInputTokens = 100
		existing.SettleUnits = 42

		incoming := newIncoming()
		view := &redisTaskView{batchID: "b-1", status: "in_progress"} // 无 usage/settle 字段
		merged := mergeTaskForSync(existing, incoming, view, now)
		assert.Equal(t, int64(100), merged.UsageInputTokens, "hash 未携带不得抹零")
		assert.Equal(t, int64(42), merged.SettleUnits)

		usageIn := int64(7)
		settle := int64(9)
		view2 := &redisTaskView{batchID: "b-1", status: "in_progress", usageIn: &usageIn, settleUnits: &settle}
		merged = mergeTaskForSync(existing, incoming, view2, now)
		assert.Equal(t, int64(7), merged.UsageInputTokens)
		assert.Equal(t, int64(9), merged.SettleUnits)
	})

	t.Run("归属列不回退", func(t *testing.T) {
		existing := newIncoming()
		existing.ID = 8
		existing.ProductName = "p1"
		existing.EntityID = "e-1"

		incoming := newIncoming() // Redis 侧 api key 已删除，归属为空
		incoming.APIKeyID = ""
		incoming.ProductName = ""
		incoming.EntityID = ""
		view := &redisTaskView{batchID: "b-1", status: "in_progress"}
		merged := mergeTaskForSync(existing, incoming, view, now)
		assert.Equal(t, "ak", merged.APIKeyID)
		assert.Equal(t, "p1", merged.ProductName)
		assert.Equal(t, "e-1", merged.EntityID)
	})
}

func TestMergeFileForSync(t *testing.T) {
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)

	t.Run("insert 语义补 seen 时间", func(t *testing.T) {
		incoming := &BatchFile{FileID: "f-1", Provider: "p", Direction: FileDirectionInput}
		merged := mergeFileForSync(nil, incoming, &redisFileView{partition: "p", fileID: "f-1"}, now)
		require.NotNil(t, merged.FirstSeenAt)
		require.NotNil(t, merged.LastSeenAt)
	})

	t.Run("update 语义刷新 last_seen 保留 first_seen", func(t *testing.T) {
		first := now.Add(-time.Hour)
		existing := &BatchFile{ID: 1, FileID: "f-1", Provider: "p", FirstSeenAt: &first, Lines: 9}
		incoming := &BatchFile{FileID: "f-1", Provider: "p"}
		merged := mergeFileForSync(existing, incoming, &redisFileView{partition: "p", fileID: "f-1"}, now)
		assert.Equal(t, first, *merged.FirstSeenAt)
		assert.Equal(t, now, *merged.LastSeenAt)
		assert.Equal(t, int64(9), merged.Lines, "hash 未携带 lines 不得抹零")
		assert.Equal(t, int64(1), merged.ID)
	})
}
