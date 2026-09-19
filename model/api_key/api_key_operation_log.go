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

package api_key

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rainway-ai-gateway/ai-gateway-api/model/ioperlog"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/shared"
)

func (rppm *APIKeyManager) recordAPIKeyOperation(ctx context.Context, action string, apiKey *APIKeyParam, before, after map[string]interface{}, err error) {
	if rppm.operationLogManager == nil || apiKey == nil {
		return
	}

	resourceID := ""
	if apiKey.ID != nil {
		resourceID = *apiKey.ID
	}

	resourceName := resourceID
	if apiKey.Description != nil && *apiKey.Description != "" {
		resourceName = *apiKey.Description
	}

	resourceParentID := ""
	if apiKey.EntityID != nil {
		resourceParentID = *apiKey.EntityID
	}

	status := ioperlog.StatusSuccess
	errorMsg := ""
	if err != nil {
		status = ioperlog.StatusFailed
		errorMsg = ioperlog.TruncateErrorMessageDefault(err)
		if apiKey.Key != nil && *apiKey.Key != "" {
			errorMsg = ioperlog.MaskErrorMessage(errorMsg, *apiKey.Key)
		}
	}

	entry := &ioperlog.OperationLogEntry{
		Action:           action,
		ResourceType:     string(ioperlog.ResourceTypeAPIKey),
		ResourceID:       resourceID,
		ResourceName:     resourceName,
		ResourceParentID: resourceParentID,
		Status:           status,
		ErrorMsg:         errorMsg,
		CreatedAt:        time.Now(),
	}

	entry.ChangeSummary = ioperlog.BuildChangeSummary(action, before, after)

	rppm.operationLogManager.Record(ctx, entry)
}

func apiKeyParamToMap(param *APIKeyParam) map[string]interface{} {
	if param == nil {
		return nil
	}

	data, err := json.Marshal(param)
	if err != nil {
		return nil
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}

	return m
}

// apiKeyResourceOwner builds the ResourceOwner for nested-resource audit
// entries owned by an API Key (issue #161).
func apiKeyResourceOwner(apiKeyID *string) shared.ResourceOwner {
	owner := shared.ResourceOwner{Type: string(ioperlog.ResourceTypeAPIKey)}
	if apiKeyID != nil {
		owner.ID = *apiKeyID
	}
	return owner
}

// auditNestedQuotaPlanCreate records the nested quota-plan create audit on the
// API key create path. planID is 0 on failure (the rolled-back ID is dropped).
func (rppm *APIKeyManager) auditNestedQuotaPlanCreate(ctx context.Context, param *APIKeyParam, planID int64, err error) {
	if rppm.quotaPlanAuditor == nil || param == nil || param.QuotaPlan == nil {
		return
	}
	rppm.quotaPlanAuditor.AuditQuotaPlanCreate(ctx, param.QuotaPlan, planID, apiKeyResourceOwner(param.ID), err)
}

// auditNestedQuotaPlanChange records the nested quota-plan audit on the API
// key update path: update when the key already had a plan, create otherwise.
func (rppm *APIKeyManager) auditNestedQuotaPlanChange(ctx context.Context, oldAPIKey, param *APIKeyParam, createdPlanID int64, oldPlan *shared.QuotaPlanParam, err error) {
	if rppm.quotaPlanAuditor == nil || param == nil || param.QuotaPlan == nil {
		return
	}
	owner := apiKeyResourceOwner(nil)
	if oldAPIKey != nil {
		owner = apiKeyResourceOwner(oldAPIKey.ID)
	}
	if oldAPIKey != nil && oldAPIKey.QuotaPlanID != nil {
		rppm.quotaPlanAuditor.AuditQuotaPlanUpdate(ctx, oldPlan, param.QuotaPlan, *oldAPIKey.QuotaPlanID, owner, err)
		return
	}
	rppm.quotaPlanAuditor.AuditQuotaPlanCreate(ctx, param.QuotaPlan, createdPlanID, owner, err)
}

// auditNestedQuotaPlanDelete records the nested quota-plan delete audit on the
// API key delete path.
func (rppm *APIKeyManager) auditNestedQuotaPlanDelete(ctx context.Context, oldAPIKey *APIKeyParam, oldPlan *shared.QuotaPlanParam, err error) {
	if rppm.quotaPlanAuditor == nil || oldAPIKey == nil || oldAPIKey.QuotaPlanID == nil {
		return
	}
	rppm.quotaPlanAuditor.AuditQuotaPlanDelete(ctx, oldPlan, *oldAPIKey.QuotaPlanID, apiKeyResourceOwner(oldAPIKey.ID), err)
}

// auditNestedRateLimitPolicyCreate records the nested rate-limit-policy
// create audit on the API key create path.
func (rppm *APIKeyManager) auditNestedRateLimitPolicyCreate(ctx context.Context, param *APIKeyParam, policyID int64, err error) {
	if rppm.rateLimitPolicyAuditor == nil || param == nil || param.RateLimitPolicy == nil {
		return
	}
	rppm.rateLimitPolicyAuditor.AuditRateLimitPolicyCreate(ctx, param.RateLimitPolicy, policyID, apiKeyResourceOwner(param.ID), err)
}

// auditNestedRateLimitPolicyChange records the nested rate-limit-policy audit
// on the API key update path.
func (rppm *APIKeyManager) auditNestedRateLimitPolicyChange(ctx context.Context, oldAPIKey, param *APIKeyParam, createdPolicyID int64, oldPolicy *shared.RateLimitPolicyParam, err error) {
	if rppm.rateLimitPolicyAuditor == nil || param == nil || param.RateLimitPolicy == nil {
		return
	}
	owner := apiKeyResourceOwner(nil)
	if oldAPIKey != nil {
		owner = apiKeyResourceOwner(oldAPIKey.ID)
	}
	if oldAPIKey != nil && oldAPIKey.RateLimitPolicyID != nil {
		rppm.rateLimitPolicyAuditor.AuditRateLimitPolicyUpdate(ctx, oldPolicy, param.RateLimitPolicy, *oldAPIKey.RateLimitPolicyID, owner, err)
		return
	}
	rppm.rateLimitPolicyAuditor.AuditRateLimitPolicyCreate(ctx, param.RateLimitPolicy, createdPolicyID, owner, err)
}

// auditNestedRateLimitPolicyDelete records the nested rate-limit-policy
// delete audit on the API key delete path.
func (rppm *APIKeyManager) auditNestedRateLimitPolicyDelete(ctx context.Context, oldAPIKey *APIKeyParam, oldPolicy *shared.RateLimitPolicyParam, err error) {
	if rppm.rateLimitPolicyAuditor == nil || oldAPIKey == nil || oldAPIKey.RateLimitPolicyID == nil {
		return
	}
	rppm.rateLimitPolicyAuditor.AuditRateLimitPolicyDelete(ctx, oldPolicy, *oldAPIKey.RateLimitPolicyID, apiKeyResourceOwner(oldAPIKey.ID), err)
}
