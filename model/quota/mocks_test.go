// Copyright(c) 2026 The Infinity AI Gateway Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package quota

import (
	"context"

	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/itxn"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iversion_control"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/shared"
)

// fakeTxn 是一个不开启真实事务的 TxnStorager mock，
// 直接把回调里的逻辑在当前上下文执行。
type fakeTxn struct{}

func (f *fakeTxn) AtomExecute(ctx context.Context, do func(context.Context) error) error {
	return do(ctx)
}

var _ itxn.TxnStorager = (*fakeTxn)(nil)

// fakeEntityTypeStorager 实现 quota.EntityTypeStorager
type fakeEntityTypeStorager struct {
	createFn func(ctx context.Context, param *EntityTypeParam) (int64, error)
	fetchFn  func(ctx context.Context, filter *EntityTypeFilter) (*EntityTypeParam, error)
	listFn   func(ctx context.Context, filter *EntityTypeFilter) ([]*EntityTypeParam, error)
	updateFn func(ctx context.Context, filter *EntityTypeFilter, param *EntityTypeParam) (int64, error)
	deleteFn func(ctx context.Context, filter *EntityTypeFilter) error

	created []*EntityTypeParam
	fetched []*EntityTypeFilter
	listed  []*EntityTypeFilter
	updated []updateEntityTypeCall
	deleted []*EntityTypeFilter
}

type updateEntityTypeCall struct {
	filter *EntityTypeFilter
	param  *EntityTypeParam
}

func (s *fakeEntityTypeStorager) CreateEntityType(ctx context.Context, param *EntityTypeParam) (int64, error) {
	s.created = append(s.created, param)
	if s.createFn != nil {
		return s.createFn(ctx, param)
	}
	return 0, nil
}

func (s *fakeEntityTypeStorager) FetchEntityType(ctx context.Context, filter *EntityTypeFilter) (*EntityTypeParam, error) {
	s.fetched = append(s.fetched, filter)
	if s.fetchFn != nil {
		return s.fetchFn(ctx, filter)
	}
	return nil, nil
}

func (s *fakeEntityTypeStorager) FetchEntityTypeList(ctx context.Context, filter *EntityTypeFilter) ([]*EntityTypeParam, error) {
	s.listed = append(s.listed, filter)
	if s.listFn != nil {
		return s.listFn(ctx, filter)
	}
	return nil, nil
}

func (s *fakeEntityTypeStorager) UpdateEntityType(ctx context.Context, filter *EntityTypeFilter, param *EntityTypeParam) (int64, error) {
	s.updated = append(s.updated, updateEntityTypeCall{filter: filter, param: param})
	if s.updateFn != nil {
		return s.updateFn(ctx, filter, param)
	}
	return 0, nil
}

func (s *fakeEntityTypeStorager) DeleteEntityType(ctx context.Context, filter *EntityTypeFilter) error {
	s.deleted = append(s.deleted, filter)
	if s.deleteFn != nil {
		return s.deleteFn(ctx, filter)
	}
	return nil
}

var _ EntityTypeStorager = (*fakeEntityTypeStorager)(nil)

// fakeEntityStorager 实现 quota.EntityStorager
type fakeEntityStorager struct {
	createFn func(ctx context.Context, param *EntityParam) (int64, error)
	fetchFn  func(ctx context.Context, filter *EntityFilter) (*EntityParam, error)
	listFn   func(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error)
	updateFn func(ctx context.Context, filter *EntityFilter, param *EntityParam) (int64, error)
	deleteFn func(ctx context.Context, filter *EntityFilter) error

	created []*EntityParam
	fetched []*EntityFilter
	listed  []*EntityFilter
	updated []updateEntityCall
	deleted []*EntityFilter
}

type updateEntityCall struct {
	filter *EntityFilter
	param  *EntityParam
}

func (s *fakeEntityStorager) CreateEntity(ctx context.Context, param *EntityParam) (int64, error) {
	s.created = append(s.created, param)
	if s.createFn != nil {
		return s.createFn(ctx, param)
	}
	return 0, nil
}

func (s *fakeEntityStorager) FetchEntity(ctx context.Context, filter *EntityFilter) (*EntityParam, error) {
	s.fetched = append(s.fetched, filter)
	if s.fetchFn != nil {
		return s.fetchFn(ctx, filter)
	}
	return nil, nil
}

func (s *fakeEntityStorager) FetchEntityList(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error) {
	s.listed = append(s.listed, filter)
	if s.listFn != nil {
		return s.listFn(ctx, filter)
	}
	return nil, nil
}

func (s *fakeEntityStorager) UpdateEntity(ctx context.Context, filter *EntityFilter, param *EntityParam) (int64, error) {
	s.updated = append(s.updated, updateEntityCall{filter: filter, param: param})
	if s.updateFn != nil {
		return s.updateFn(ctx, filter, param)
	}
	return 0, nil
}

func (s *fakeEntityStorager) DeleteEntity(ctx context.Context, filter *EntityFilter) error {
	s.deleted = append(s.deleted, filter)
	if s.deleteFn != nil {
		return s.deleteFn(ctx, filter)
	}
	return nil
}

var _ EntityStorager = (*fakeEntityStorager)(nil)

// fakeQuotaPlanStorager 实现 quota.QuotaPlanStorager（内部类型）
type fakeQuotaPlanStorager struct {
	createFn func(ctx context.Context, param *QuotaPlanParam) (int64, error)
	fetchFn  func(ctx context.Context, filter *QuotaPlanFilter) (*QuotaPlanParam, error)
	listFn   func(ctx context.Context, filter *QuotaPlanFilter) ([]*QuotaPlanParam, error)
	updateFn func(ctx context.Context, filter *QuotaPlanFilter, param *QuotaPlanParam) (int64, error)
	deleteFn func(ctx context.Context, filter *QuotaPlanFilter) error

	created []*QuotaPlanParam
	fetched []*QuotaPlanFilter
	listed  []*QuotaPlanFilter
	updated []updateQuotaPlanCall
	deleted []*QuotaPlanFilter
}

type updateQuotaPlanCall struct {
	filter *QuotaPlanFilter
	param  *QuotaPlanParam
}

func (s *fakeQuotaPlanStorager) CreateQuotaPlan(ctx context.Context, param *QuotaPlanParam) (int64, error) {
	s.created = append(s.created, param)
	if s.createFn != nil {
		return s.createFn(ctx, param)
	}
	return 0, nil
}

func (s *fakeQuotaPlanStorager) FetchQuotaPlan(ctx context.Context, filter *QuotaPlanFilter) (*QuotaPlanParam, error) {
	s.fetched = append(s.fetched, filter)
	if s.fetchFn != nil {
		return s.fetchFn(ctx, filter)
	}
	return nil, nil
}

func (s *fakeQuotaPlanStorager) FetchQuotaPlanList(ctx context.Context, filter *QuotaPlanFilter) ([]*QuotaPlanParam, error) {
	s.listed = append(s.listed, filter)
	if s.listFn != nil {
		return s.listFn(ctx, filter)
	}
	return nil, nil
}

func (s *fakeQuotaPlanStorager) UpdateQuotaPlan(ctx context.Context, filter *QuotaPlanFilter, param *QuotaPlanParam) (int64, error) {
	s.updated = append(s.updated, updateQuotaPlanCall{filter: filter, param: param})
	if s.updateFn != nil {
		return s.updateFn(ctx, filter, param)
	}
	return 0, nil
}

func (s *fakeQuotaPlanStorager) DeleteQuotaPlan(ctx context.Context, filter *QuotaPlanFilter) error {
	s.deleted = append(s.deleted, filter)
	if s.deleteFn != nil {
		return s.deleteFn(ctx, filter)
	}
	return nil
}

var _ QuotaPlanStorager = (*fakeQuotaPlanStorager)(nil)

// fakeSharedQuotaPlanStorager 实现 shared.QuotaPlanStorager
type fakeSharedQuotaPlanStorager struct {
	createFn func(ctx context.Context, param *shared.QuotaPlanParam) (int64, error)
	updateFn func(ctx context.Context, id int64, param *shared.QuotaPlanParam) (int64, error)
	deleteFn func(ctx context.Context, id int64) error
	fetchFn  func(ctx context.Context, id int64) (*shared.QuotaPlanParam, error)

	created []*shared.QuotaPlanParam
	updated []updateSharedQuotaPlanCall
	deleted []int64
	fetched []int64
}

type updateSharedQuotaPlanCall struct {
	id    int64
	param *shared.QuotaPlanParam
}

func (s *fakeSharedQuotaPlanStorager) CreateQuotaPlan(ctx context.Context, param *shared.QuotaPlanParam) (int64, error) {
	s.created = append(s.created, param)
	if s.createFn != nil {
		return s.createFn(ctx, param)
	}
	return 0, nil
}

func (s *fakeSharedQuotaPlanStorager) UpdateQuotaPlan(ctx context.Context, id int64, param *shared.QuotaPlanParam) (int64, error) {
	s.updated = append(s.updated, updateSharedQuotaPlanCall{id: id, param: param})
	if s.updateFn != nil {
		return s.updateFn(ctx, id, param)
	}
	return 0, nil
}

func (s *fakeSharedQuotaPlanStorager) DeleteQuotaPlan(ctx context.Context, id int64) error {
	s.deleted = append(s.deleted, id)
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}
	return nil
}

func (s *fakeSharedQuotaPlanStorager) FetchQuotaPlan(ctx context.Context, id int64) (*shared.QuotaPlanParam, error) {
	s.fetched = append(s.fetched, id)
	if s.fetchFn != nil {
		return s.fetchFn(ctx, id)
	}
	return nil, nil
}

var _ shared.QuotaPlanStorager = (*fakeSharedQuotaPlanStorager)(nil)

// fakeRateLimitPolicyStorager 实现 quota.RateLimitPolicyStorager（内部类型）
type fakeRateLimitPolicyStorager struct {
	createFn func(ctx context.Context, param *RateLimitPolicyParam) (int64, error)
	fetchFn  func(ctx context.Context, filter *RateLimitPolicyFilter) (*RateLimitPolicyParam, error)
	listFn   func(ctx context.Context, filter *RateLimitPolicyFilter) ([]*RateLimitPolicyParam, error)
	updateFn func(ctx context.Context, filter *RateLimitPolicyFilter, param *RateLimitPolicyParam) (int64, error)
	deleteFn func(ctx context.Context, filter *RateLimitPolicyFilter) error

	created []*RateLimitPolicyParam
	fetched []*RateLimitPolicyFilter
	listed  []*RateLimitPolicyFilter
	updated []updateRateLimitPolicyCall
	deleted []*RateLimitPolicyFilter
}

type updateRateLimitPolicyCall struct {
	filter *RateLimitPolicyFilter
	param  *RateLimitPolicyParam
}

func (s *fakeRateLimitPolicyStorager) CreateRateLimitPolicy(ctx context.Context, param *RateLimitPolicyParam) (int64, error) {
	s.created = append(s.created, param)
	if s.createFn != nil {
		return s.createFn(ctx, param)
	}
	return 0, nil
}

func (s *fakeRateLimitPolicyStorager) FetchRateLimitPolicy(ctx context.Context, filter *RateLimitPolicyFilter) (*RateLimitPolicyParam, error) {
	s.fetched = append(s.fetched, filter)
	if s.fetchFn != nil {
		return s.fetchFn(ctx, filter)
	}
	return nil, nil
}

func (s *fakeRateLimitPolicyStorager) FetchRateLimitPolicyList(ctx context.Context, filter *RateLimitPolicyFilter) ([]*RateLimitPolicyParam, error) {
	s.listed = append(s.listed, filter)
	if s.listFn != nil {
		return s.listFn(ctx, filter)
	}
	return nil, nil
}

func (s *fakeRateLimitPolicyStorager) UpdateRateLimitPolicy(ctx context.Context, filter *RateLimitPolicyFilter, param *RateLimitPolicyParam) (int64, error) {
	s.updated = append(s.updated, updateRateLimitPolicyCall{filter: filter, param: param})
	if s.updateFn != nil {
		return s.updateFn(ctx, filter, param)
	}
	return 0, nil
}

func (s *fakeRateLimitPolicyStorager) DeleteRateLimitPolicy(ctx context.Context, filter *RateLimitPolicyFilter) error {
	s.deleted = append(s.deleted, filter)
	if s.deleteFn != nil {
		return s.deleteFn(ctx, filter)
	}
	return nil
}

var _ RateLimitPolicyStorager = (*fakeRateLimitPolicyStorager)(nil)

// fakeSharedRateLimitPolicyStorager 实现 shared.RateLimitPolicyStorager
type fakeSharedRateLimitPolicyStorager struct {
	createFn func(ctx context.Context, param *shared.RateLimitPolicyParam) (int64, error)
	updateFn func(ctx context.Context, id int64, param *shared.RateLimitPolicyParam) (int64, error)
	deleteFn func(ctx context.Context, id int64) error
	fetchFn  func(ctx context.Context, id int64) (*shared.RateLimitPolicyParam, error)

	created []*shared.RateLimitPolicyParam
	updated []updateSharedRateLimitPolicyCall
	deleted []int64
	fetched []int64
}

type updateSharedRateLimitPolicyCall struct {
	id    int64
	param *shared.RateLimitPolicyParam
}

func (s *fakeSharedRateLimitPolicyStorager) CreateRateLimitPolicy(ctx context.Context, param *shared.RateLimitPolicyParam) (int64, error) {
	s.created = append(s.created, param)
	if s.createFn != nil {
		return s.createFn(ctx, param)
	}
	return 0, nil
}

func (s *fakeSharedRateLimitPolicyStorager) UpdateRateLimitPolicy(ctx context.Context, id int64, param *shared.RateLimitPolicyParam) (int64, error) {
	s.updated = append(s.updated, updateSharedRateLimitPolicyCall{id: id, param: param})
	if s.updateFn != nil {
		return s.updateFn(ctx, id, param)
	}
	return 0, nil
}

func (s *fakeSharedRateLimitPolicyStorager) DeleteRateLimitPolicy(ctx context.Context, id int64) error {
	s.deleted = append(s.deleted, id)
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}
	return nil
}

func (s *fakeSharedRateLimitPolicyStorager) FetchRateLimitPolicy(ctx context.Context, id int64) (*shared.RateLimitPolicyParam, error) {
	s.fetched = append(s.fetched, id)
	if s.fetchFn != nil {
		return s.fetchFn(ctx, id)
	}
	return nil, nil
}

var _ shared.RateLimitPolicyStorager = (*fakeSharedRateLimitPolicyStorager)(nil)

// fakeQuotaBalanceStorager 实现 quota.QuotaBalanceStorager
type fakeQuotaBalanceStorager struct {
	createFn func(ctx context.Context, param *QuotaBalanceParam) (int64, error)
	fetchFn  func(ctx context.Context, filter *QuotaBalanceFilter) (*QuotaBalanceParam, error)
	listFn   func(ctx context.Context, filter *QuotaBalanceFilter) ([]*QuotaBalanceParam, error)
	updateFn func(ctx context.Context, filter *QuotaBalanceFilter, param *QuotaBalanceParam) (int64, error)
	deleteFn func(ctx context.Context, filter *QuotaBalanceFilter) error

	created []*QuotaBalanceParam
	fetched []*QuotaBalanceFilter
	listed  []*QuotaBalanceFilter
	updated []updateQuotaBalanceCall
	deleted []*QuotaBalanceFilter
}

type updateQuotaBalanceCall struct {
	filter *QuotaBalanceFilter
	param  *QuotaBalanceParam
}

func (s *fakeQuotaBalanceStorager) CreateQuotaBalance(ctx context.Context, param *QuotaBalanceParam) (int64, error) {
	s.created = append(s.created, param)
	if s.createFn != nil {
		return s.createFn(ctx, param)
	}
	return 0, nil
}

func (s *fakeQuotaBalanceStorager) FetchQuotaBalance(ctx context.Context, filter *QuotaBalanceFilter) (*QuotaBalanceParam, error) {
	s.fetched = append(s.fetched, filter)
	if s.fetchFn != nil {
		return s.fetchFn(ctx, filter)
	}
	return nil, nil
}

func (s *fakeQuotaBalanceStorager) FetchQuotaBalanceList(ctx context.Context, filter *QuotaBalanceFilter) ([]*QuotaBalanceParam, error) {
	s.listed = append(s.listed, filter)
	if s.listFn != nil {
		return s.listFn(ctx, filter)
	}
	return nil, nil
}

func (s *fakeQuotaBalanceStorager) UpdateQuotaBalance(ctx context.Context, filter *QuotaBalanceFilter, param *QuotaBalanceParam) (int64, error) {
	s.updated = append(s.updated, updateQuotaBalanceCall{filter: filter, param: param})
	if s.updateFn != nil {
		return s.updateFn(ctx, filter, param)
	}
	return 0, nil
}

func (s *fakeQuotaBalanceStorager) DeleteQuotaBalance(ctx context.Context, filter *QuotaBalanceFilter) error {
	s.deleted = append(s.deleted, filter)
	if s.deleteFn != nil {
		return s.deleteFn(ctx, filter)
	}
	return nil
}

var _ QuotaBalanceStorager = (*fakeQuotaBalanceStorager)(nil)

// fakeRouteRulesStorager 实现 shared.RouteRulesStorager
type fakeRouteRulesStorager struct {
	createFn func(ctx context.Context, ruleType string, owner *string, param *shared.RouteRulesParam) (int64, error)
	fetchFn  func(ctx context.Context, ruleType string, owner *string) (*shared.RouteRulesParam, error)
	listFn   func(ctx context.Context, filter *shared.RouteRulesFilter) ([]*shared.RouteTableParam, int64, error)
	updateFn func(ctx context.Context, id int64, param *shared.RouteRulesParam) (int64, error)
	deleteFn func(ctx context.Context, id int64) error
	fetchByIDFn func(ctx context.Context, id int64) (*shared.RouteRulesParam, error)

	created     []createRouteRulesCall
	fetched     []fetchRouteRulesCall
	listed      []*shared.RouteRulesFilter
	updated     []updateRouteRulesCall
	deleted     []int64
	fetchedByID []int64
}

type createRouteRulesCall struct {
	ruleType string
	owner    *string
	param    *shared.RouteRulesParam
}

type fetchRouteRulesCall struct {
	ruleType string
	owner    *string
}

type updateRouteRulesCall struct {
	id    int64
	param *shared.RouteRulesParam
}

func (s *fakeRouteRulesStorager) CreateRouteRules(ctx context.Context, ruleType string, owner *string, param *shared.RouteRulesParam) (int64, error) {
	s.created = append(s.created, createRouteRulesCall{ruleType: ruleType, owner: owner, param: param})
	if s.createFn != nil {
		return s.createFn(ctx, ruleType, owner, param)
	}
	return 0, nil
}

func (s *fakeRouteRulesStorager) FetchRouteRules(ctx context.Context, ruleType string, owner *string) (*shared.RouteRulesParam, error) {
	s.fetched = append(s.fetched, fetchRouteRulesCall{ruleType: ruleType, owner: owner})
	if s.fetchFn != nil {
		return s.fetchFn(ctx, ruleType, owner)
	}
	return nil, nil
}

func (s *fakeRouteRulesStorager) FetchRouteRulesList(ctx context.Context, filter *shared.RouteRulesFilter) ([]*shared.RouteTableParam, int64, error) {
	s.listed = append(s.listed, filter)
	if s.listFn != nil {
		return s.listFn(ctx, filter)
	}
	return nil, 0, nil
}

func (s *fakeRouteRulesStorager) UpdateRouteRules(ctx context.Context, id int64, param *shared.RouteRulesParam) (int64, error) {
	s.updated = append(s.updated, updateRouteRulesCall{id: id, param: param})
	if s.updateFn != nil {
		return s.updateFn(ctx, id, param)
	}
	return 0, nil
}

func (s *fakeRouteRulesStorager) DeleteRouteRules(ctx context.Context, id int64) error {
	s.deleted = append(s.deleted, id)
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}
	return nil
}

func (s *fakeRouteRulesStorager) FetchRouteRulesByID(ctx context.Context, id int64) (*shared.RouteRulesParam, error) {
	s.fetchedByID = append(s.fetchedByID, id)
	if s.fetchByIDFn != nil {
		return s.fetchByIDFn(ctx, id)
	}
	return nil, nil
}

var _ shared.RouteRulesStorager = (*fakeRouteRulesStorager)(nil)

// fakeAPIKeyStorager 实现 icluster_conf.APIKeyStorager
type fakeAPIKeyStorager struct {
	fetchListFn      func(ctx context.Context, filter *icluster_conf.APIKeyFilter) ([]*icluster_conf.APIKeyParam, error)
	createFn         func(ctx context.Context, param *icluster_conf.APIKeyParam) (int64, error)
	updateFn         func(ctx context.Context, filter *icluster_conf.APIKeyFilter, param *icluster_conf.APIKeyParam) (int64, error)
	deleteFn         func(ctx context.Context, filter *icluster_conf.APIKeyFilter) error
	createTokenFn    func(ctx context.Context, param *icluster_conf.APIKeyTokenParam) (int64, error)
	updateTokenFn    func(ctx context.Context, filter *icluster_conf.APIKeyTokenFilter, param *icluster_conf.APIKeyTokenParam) error
	fetchTokenListFn func(ctx context.Context, filter *icluster_conf.APIKeyTokenFilter) ([]*icluster_conf.APIKeyTokenParam, error)
}

func (s *fakeAPIKeyStorager) FetchAPIKeyList(ctx context.Context, filter *icluster_conf.APIKeyFilter) ([]*icluster_conf.APIKeyParam, error) {
	if s.fetchListFn != nil {
		return s.fetchListFn(ctx, filter)
	}
	return nil, nil
}

func (s *fakeAPIKeyStorager) CreateAPIKey(ctx context.Context, param *icluster_conf.APIKeyParam) (int64, error) {
	if s.createFn != nil {
		return s.createFn(ctx, param)
	}
	return 0, nil
}

func (s *fakeAPIKeyStorager) UpdateAPIKey(ctx context.Context, filter *icluster_conf.APIKeyFilter, param *icluster_conf.APIKeyParam) (int64, error) {
	if s.updateFn != nil {
		return s.updateFn(ctx, filter, param)
	}
	return 0, nil
}

func (s *fakeAPIKeyStorager) DeleteAPIKey(ctx context.Context, filter *icluster_conf.APIKeyFilter) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, filter)
	}
	return nil
}

func (s *fakeAPIKeyStorager) CreateAPIKeyToken(ctx context.Context, param *icluster_conf.APIKeyTokenParam) (int64, error) {
	if s.createTokenFn != nil {
		return s.createTokenFn(ctx, param)
	}
	return 0, nil
}

func (s *fakeAPIKeyStorager) UpdateAPIKeyToken(ctx context.Context, filter *icluster_conf.APIKeyTokenFilter, param *icluster_conf.APIKeyTokenParam) error {
	if s.updateTokenFn != nil {
		return s.updateTokenFn(ctx, filter, param)
	}
	return nil
}

func (s *fakeAPIKeyStorager) FetchAPIKeyTokenList(ctx context.Context, filter *icluster_conf.APIKeyTokenFilter) ([]*icluster_conf.APIKeyTokenParam, error) {
	if s.fetchTokenListFn != nil {
		return s.fetchTokenListFn(ctx, filter)
	}
	return nil, nil
}

var _ icluster_conf.APIKeyStorager = (*fakeAPIKeyStorager)(nil)

// fakeVersionControlStorager 实现 iversion_control.VersionControlStorager
type fakeVersionControlStorager struct {
	upsertFn func(ctx context.Context, css *iversion_control.ExportData) (string, error)
}

func (s *fakeVersionControlStorager) UpsertConfigLastExportedVersion(ctx context.Context, css *iversion_control.ExportData) (string, error) {
	if s.upsertFn != nil {
		return s.upsertFn(ctx, css)
	}
	return "", nil
}

var _ iversion_control.VersionControlStorager = (*fakeVersionControlStorager)(nil)
