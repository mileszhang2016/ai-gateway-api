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
//limitations under the License.

package quota

import (
	"context"
	"strconv"
	"time"

	"github.com/rainway-ai-gateway/ai-gateway-api/model/ioperlog"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/shared"
)

func (m *QuotaPlanManager) recordQuotaPlanOperation(ctx context.Context, action string, planID, parentID string, before, after map[string]interface{}, err error) {
	if m.operationLogManager == nil {
		return
	}

	status := ioperlog.StatusSuccess
	errorMsg := ""
	if err != nil {
		status = ioperlog.StatusFailed
		errorMsg = ioperlog.TruncateErrorMessageDefault(err)
	}

	entry := &ioperlog.OperationLogEntry{
		Action:           action,
		ResourceType:     string(ioperlog.ResourceTypeQuotaPlan),
		ResourceID:       planID,
		ResourceName:     "",
		ResourceParentID: parentID,
		Status:           status,
		ErrorMsg:         errorMsg,
		CreatedAt:        time.Now(),
	}

	entry.ChangeSummary = ioperlog.BuildChangeSummary(action, before, after)

	m.operationLogManager.Record(ctx, entry)
}

func quotaPlanIDString(id int64) string {
	return strconv.FormatInt(id, 10)
}

// AuditQuotaPlanCreate records the audit entry for a quota-plan create whose
// write was executed inside the owning resource's transaction (issue #161).
// planID is the created plan ID on success and 0 on failure.
func (m *QuotaPlanManager) AuditQuotaPlanCreate(ctx context.Context, param *shared.QuotaPlanParam, planID int64, owner shared.ResourceOwner, err error) {
	resourceID := ""
	if planID > 0 {
		resourceID = quotaPlanIDString(planID)
	}
	m.recordQuotaPlanOperation(ctx, string(ioperlog.ActionCreate), resourceID, owner.ID, nil, quotaPlanParamToMap(quotaPlanParamFromShared(param)), err)
}

// AuditQuotaPlanUpdate records the audit entry for a quota-plan update whose
// write was executed inside the owning resource's transaction.
func (m *QuotaPlanManager) AuditQuotaPlanUpdate(ctx context.Context, oldPlan, param *shared.QuotaPlanParam, planID int64, owner shared.ResourceOwner, err error) {
	resourceID := ""
	if planID > 0 {
		resourceID = quotaPlanIDString(planID)
	}
	m.recordQuotaPlanOperation(ctx, string(ioperlog.ActionUpdate), resourceID, owner.ID, quotaPlanParamToMap(quotaPlanParamFromShared(oldPlan)), quotaPlanParamToMap(quotaPlanParamFromShared(param)), err)
}

// AuditQuotaPlanDelete records the audit entry for a quota-plan delete whose
// write was executed inside the owning resource's transaction.
func (m *QuotaPlanManager) AuditQuotaPlanDelete(ctx context.Context, oldPlan *shared.QuotaPlanParam, planID int64, owner shared.ResourceOwner, err error) {
	resourceID := ""
	if planID > 0 {
		resourceID = quotaPlanIDString(planID)
	}
	m.recordQuotaPlanOperation(ctx, string(ioperlog.ActionDelete), resourceID, owner.ID, quotaPlanParamToMap(quotaPlanParamFromShared(oldPlan)), nil, err)
}

// quotaPlanParamFromShared converts the shared quota-plan param into the
// storage-level param so the audit map builder can be reused.
func quotaPlanParamFromShared(param *shared.QuotaPlanParam) *QuotaPlanParam {
	if param == nil {
		return nil
	}
	return &QuotaPlanParam{
		Unlimited:             param.Unlimited,
		PassWhenNoEnoughQuota: param.PassWhenNoEnoughQuota,
		Quota:                 param.Quota,
		Unit:                  param.Unit,
		ResetPeriod:           param.ResetPeriod,
		LastResetAt:           param.LastResetAt,
	}
}

func quotaPlanParamToMap(param *QuotaPlanParam) map[string]interface{} {
	if param == nil {
		return nil
	}

	m := map[string]interface{}{}
	if param.ID != nil {
		m["id"] = *param.ID
	}
	if param.Unlimited != nil {
		m["unlimited"] = *param.Unlimited
	}
	if param.PassWhenNoEnoughQuota != nil {
		m["pass_when_no_enough_quota"] = *param.PassWhenNoEnoughQuota
	}
	if param.Quota != nil {
		m["quota"] = *param.Quota
	}
	if param.Unit != nil {
		m["unit"] = *param.Unit
	}
	if param.ResetPeriod != nil {
		m["reset_period"] = *param.ResetPeriod
	}
	if param.LastResetAt != nil {
		m["last_reset_at"] = *param.LastResetAt
	}

	return m
}
