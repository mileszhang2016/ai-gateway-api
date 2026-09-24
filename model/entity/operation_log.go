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

package entity

import (
	"context"
	"time"

	"github.com/rainway-ai-gateway/ai-gateway-api/model/ioperlog"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/shared"
)

func (m *EntityManager) recordEntityOperation(ctx context.Context, action string, entityID, entityName, parentID string, before, after map[string]interface{}, err error) {
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
		ResourceType:     string(ioperlog.ResourceTypeEntity),
		ResourceID:       entityID,
		ResourceName:     entityName,
		ResourceParentID: parentID,
		Status:           status,
		ErrorMsg:         errorMsg,
		CreatedAt:        time.Now(),
	}

	entry.ChangeSummary = ioperlog.BuildChangeSummary(action, before, after)

	m.operationLogManager.Record(ctx, entry)
}

func (m *EntityTypeManager) recordEntityTypeOperation(ctx context.Context, action string, param *EntityTypeParam, before, after map[string]interface{}, err error) {
	if m.operationLogManager == nil || param == nil {
		return
	}

	resourceID := ""
	if param.TypeName != nil {
		resourceID = *param.TypeName
	}
	resourceName := resourceID
	if param.Description != nil && *param.Description != "" {
		resourceName = *param.Description
	}

	status := ioperlog.StatusSuccess
	errorMsg := ""
	if err != nil {
		status = ioperlog.StatusFailed
		errorMsg = ioperlog.TruncateErrorMessageDefault(err)
	}

	entry := &ioperlog.OperationLogEntry{
		Action:       action,
		ResourceType: string(ioperlog.ResourceTypeEntityType),
		ResourceID:   resourceID,
		ResourceName: resourceName,
		Status:       status,
		ErrorMsg:     errorMsg,
		CreatedAt:    time.Now(),
	}

	entry.ChangeSummary = ioperlog.BuildChangeSummary(action, before, after)

	m.operationLogManager.Record(ctx, entry)
}

func entityTypeParamToMap(param *EntityTypeParam) map[string]interface{} {
	if param == nil {
		return nil
	}

	m := map[string]interface{}{}
	if param.TypeName != nil {
		m["type_name"] = *param.TypeName
	}
	if param.Description != nil {
		m["description"] = *param.Description
	}
	if param.Level != nil {
		m["level"] = *param.Level
	}

	return m
}

func entityParamToMap(param *EntityParam) map[string]interface{} {
	if param == nil {
		return nil
	}

	m := map[string]interface{}{}
	if param.Name != nil {
		m["name"] = *param.Name
	}
	if param.Description != nil {
		m["description"] = *param.Description
	}
	if param.Type != nil {
		m["type"] = *param.Type
	}
	if param.ParentID != nil {
		m["parent_id"] = *param.ParentID
	}
	if param.EntityID != nil {
		m["entity_id"] = *param.EntityID
	}
	if len(param.AllowModels) > 0 {
		m["allow_models"] = param.AllowModels
	}
	if len(param.BlockModels) > 0 {
		m["block_models"] = param.BlockModels
	}

	return m
}

// entityResourceOwner builds the ResourceOwner for nested-resource audit
// entries owned by an Entity (issue #161).
func entityResourceOwner(entityID *string) shared.ResourceOwner {
	owner := shared.ResourceOwner{Type: string(ioperlog.ResourceTypeEntity)}
	if entityID != nil {
		owner.ID = *entityID
	}
	return owner
}

// auditNestedQuotaPlanCreate records the nested quota-plan create audit on the
// entity create path. planID is 0 on failure (the rolled-back ID is dropped).
func (m *EntityManager) auditNestedQuotaPlanCreate(ctx context.Context, param *EntityParam, planID int64, err error) {
	if m.quotaPlanAuditor == nil || param == nil || param.QuotaPlan == nil {
		return
	}
	m.quotaPlanAuditor.AuditQuotaPlanCreate(ctx, param.QuotaPlan, planID, entityResourceOwner(param.EntityID), err)
}

// auditNestedQuotaPlanChange records the nested quota-plan audit on the entity
// update path: update when the entity already had a plan, create otherwise.
func (m *EntityManager) auditNestedQuotaPlanChange(ctx context.Context, oldEntity, param *EntityParam, createdPlanID int64, oldPlan *shared.QuotaPlanParam, err error) {
	if m.quotaPlanAuditor == nil || param == nil || param.QuotaPlan == nil {
		return
	}
	owner := entityResourceOwner(nil)
	if oldEntity != nil {
		owner = entityResourceOwner(oldEntity.EntityID)
	}
	if oldEntity != nil && oldEntity.QuotaPlanID != nil {
		m.quotaPlanAuditor.AuditQuotaPlanUpdate(ctx, oldPlan, param.QuotaPlan, *oldEntity.QuotaPlanID, owner, err)
		return
	}
	m.quotaPlanAuditor.AuditQuotaPlanCreate(ctx, param.QuotaPlan, createdPlanID, owner, err)
}

// auditNestedQuotaPlanDelete records the nested quota-plan delete audit on the
// entity delete path.
func (m *EntityManager) auditNestedQuotaPlanDelete(ctx context.Context, oldEntity *EntityParam, oldPlan *shared.QuotaPlanParam, err error) {
	if m.quotaPlanAuditor == nil || oldEntity == nil || oldEntity.QuotaPlanID == nil {
		return
	}
	m.quotaPlanAuditor.AuditQuotaPlanDelete(ctx, oldPlan, *oldEntity.QuotaPlanID, entityResourceOwner(oldEntity.EntityID), err)
}

// auditNestedRateLimitPolicyCreate records the nested rate-limit-policy
// create audit on the entity create path.
func (m *EntityManager) auditNestedRateLimitPolicyCreate(ctx context.Context, param *EntityParam, policyID int64, err error) {
	if m.rateLimitPolicyAuditor == nil || param == nil || param.RateLimitPolicy == nil {
		return
	}
	m.rateLimitPolicyAuditor.AuditRateLimitPolicyCreate(ctx, param.RateLimitPolicy, policyID, entityResourceOwner(param.EntityID), err)
}

// auditNestedRateLimitPolicyChange records the nested rate-limit-policy audit
// on the entity update path.
func (m *EntityManager) auditNestedRateLimitPolicyChange(ctx context.Context, oldEntity, param *EntityParam, createdPolicyID int64, oldPolicy *shared.RateLimitPolicyParam, err error) {
	if m.rateLimitPolicyAuditor == nil || param == nil || param.RateLimitPolicy == nil {
		return
	}
	owner := entityResourceOwner(nil)
	if oldEntity != nil {
		owner = entityResourceOwner(oldEntity.EntityID)
	}
	if oldEntity != nil && oldEntity.RateLimitPolicyID != nil {
		m.rateLimitPolicyAuditor.AuditRateLimitPolicyUpdate(ctx, oldPolicy, param.RateLimitPolicy, *oldEntity.RateLimitPolicyID, owner, err)
		return
	}
	m.rateLimitPolicyAuditor.AuditRateLimitPolicyCreate(ctx, param.RateLimitPolicy, createdPolicyID, owner, err)
}

// auditNestedRateLimitPolicyDelete records the nested rate-limit-policy
// delete audit on the entity delete path.
func (m *EntityManager) auditNestedRateLimitPolicyDelete(ctx context.Context, oldEntity *EntityParam, oldPolicy *shared.RateLimitPolicyParam, err error) {
	if m.rateLimitPolicyAuditor == nil || oldEntity == nil || oldEntity.RateLimitPolicyID == nil {
		return
	}
	m.rateLimitPolicyAuditor.AuditRateLimitPolicyDelete(ctx, oldPolicy, *oldEntity.RateLimitPolicyID, entityResourceOwner(oldEntity.EntityID), err)
}
