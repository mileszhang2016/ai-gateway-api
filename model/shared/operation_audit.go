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

package shared

import "context"

// ResourceOwner identifies the owning resource of a nested write (e.g. the
// Entity or API Key that owns a quota plan). ID is the owner's business ID
// and is recorded as resource_parent_id in operation logs.
type ResourceOwner struct {
	// Type is the owner resource type, e.g. the ioperlog resource-type value
	// of the owning resource ("entity" / "api_key"). Documentary only; the
	// operation-log entry itself only carries the ID.
	Type string
	// ID is the owner's business ID; empty when the owner is unknown.
	ID string
}

// QuotaPlanAuditor records operation logs for quota-plan writes that are
// performed inside the owning resource's transaction (nested create/update/
// delete on Entity / API Key). The write itself stays in the owner's
// transaction; implementations only build and record the audit entry, with
// err reporting the outcome of the write path (failed entries are recorded
// when err is non-nil).
type QuotaPlanAuditor interface {
	AuditQuotaPlanCreate(ctx context.Context, param *QuotaPlanParam, planID int64, owner ResourceOwner, err error)
	AuditQuotaPlanUpdate(ctx context.Context, oldPlan, param *QuotaPlanParam, planID int64, owner ResourceOwner, err error)
	AuditQuotaPlanDelete(ctx context.Context, oldPlan *QuotaPlanParam, planID int64, owner ResourceOwner, err error)
}

// RateLimitPolicyAuditor is the RateLimitPolicy counterpart of
// QuotaPlanAuditor for nested rate-limit-policy writes.
type RateLimitPolicyAuditor interface {
	AuditRateLimitPolicyCreate(ctx context.Context, param *RateLimitPolicyParam, policyID int64, owner ResourceOwner, err error)
	AuditRateLimitPolicyUpdate(ctx context.Context, oldPolicy, param *RateLimitPolicyParam, policyID int64, owner ResourceOwner, err error)
	AuditRateLimitPolicyDelete(ctx context.Context, oldPolicy *RateLimitPolicyParam, policyID int64, owner ResourceOwner, err error)
}
