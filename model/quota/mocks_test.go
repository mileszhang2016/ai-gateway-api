// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
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

	"github.com/rainway-ai-gateway/ai-gateway-api/model/api_key"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/entity"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/itxn"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iversion_control"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/rate_limit_policy"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/shared"
)

// fakeTxn 是一个不开启真实事务的 TxnStorager mock，
// 直接把回调里的逻辑在当前上下文执行。
type fakeTxn struct{}

func (f *fakeTxn) AtomExecute(ctx context.Context, do func(context.Context) error) error {
	return do(ctx)
}

var _ itxn.TxnStorager = (*fakeTxn)(nil)

// fakeEntityStorager 实现 entity.EntityStorager
type fakeEntityStorager struct {
	createFn func(ctx context.Context, param *entity.EntityParam) (int64, error)
	fetchFn  func(ctx context.Context, filter *entity.EntityFilter) (*entity.EntityParam, error)
	listFn   func(ctx context.Context, filter *entity.EntityFilter) ([]*entity.EntityParam, error)
	updateFn func(ctx context.Context, filter *entity.EntityFilter, param *entity.EntityParam) (int64, error)
	deleteFn func(ctx context.Context, filter *entity.EntityFilter) error
}

func (s *fakeEntityStorager) CreateEntity(ctx context.Context, param *entity.EntityParam) (int64, error) {
	if s.createFn != nil {
		return s.createFn(ctx, param)
	}
	return 0, nil
}

func (s *fakeEntityStorager) FetchEntity(ctx context.Context, filter *entity.EntityFilter) (*entity.EntityParam, error) {
	if s.fetchFn != nil {
		return s.fetchFn(ctx, filter)
	}
	return nil, nil
}

func (s *fakeEntityStorager) FetchEntityList(ctx context.Context, filter *entity.EntityFilter) ([]*entity.EntityParam, error) {
	if s.listFn != nil {
		return s.listFn(ctx, filter)
	}
	return nil, nil
}

func (s *fakeEntityStorager) UpdateEntity(ctx context.Context, filter *entity.EntityFilter, param *entity.EntityParam) (int64, error) {
	if s.updateFn != nil {
		return s.updateFn(ctx, filter, param)
	}
	return 0, nil
}

func (s *fakeEntityStorager) DeleteEntity(ctx context.Context, filter *entity.EntityFilter) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, filter)
	}
	return nil
}

var _ entity.EntityStorager = (*fakeEntityStorager)(nil)

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

// fakeRateLimitPolicyStorager 实现 rate_limit_policy.RateLimitPolicyStorager
type fakeRateLimitPolicyStorager struct {
	createFn func(ctx context.Context, param *rate_limit_policy.RateLimitPolicyParam) (int64, error)
	fetchFn  func(ctx context.Context, filter *rate_limit_policy.RateLimitPolicyFilter) (*rate_limit_policy.RateLimitPolicyParam, error)
	listFn   func(ctx context.Context, filter *rate_limit_policy.RateLimitPolicyFilter) ([]*rate_limit_policy.RateLimitPolicyParam, error)
	updateFn func(ctx context.Context, filter *rate_limit_policy.RateLimitPolicyFilter, param *rate_limit_policy.RateLimitPolicyParam) (int64, error)
	deleteFn func(ctx context.Context, filter *rate_limit_policy.RateLimitPolicyFilter) error
}

func (s *fakeRateLimitPolicyStorager) CreateRateLimitPolicy(ctx context.Context, param *rate_limit_policy.RateLimitPolicyParam) (int64, error) {
	if s.createFn != nil {
		return s.createFn(ctx, param)
	}
	return 0, nil
}

func (s *fakeRateLimitPolicyStorager) FetchRateLimitPolicy(ctx context.Context, filter *rate_limit_policy.RateLimitPolicyFilter) (*rate_limit_policy.RateLimitPolicyParam, error) {
	if s.fetchFn != nil {
		return s.fetchFn(ctx, filter)
	}
	return nil, nil
}

func (s *fakeRateLimitPolicyStorager) FetchRateLimitPolicyList(ctx context.Context, filter *rate_limit_policy.RateLimitPolicyFilter) ([]*rate_limit_policy.RateLimitPolicyParam, error) {
	if s.listFn != nil {
		return s.listFn(ctx, filter)
	}
	return nil, nil
}

func (s *fakeRateLimitPolicyStorager) UpdateRateLimitPolicy(ctx context.Context, filter *rate_limit_policy.RateLimitPolicyFilter, param *rate_limit_policy.RateLimitPolicyParam) (int64, error) {
	if s.updateFn != nil {
		return s.updateFn(ctx, filter, param)
	}
	return 0, nil
}

func (s *fakeRateLimitPolicyStorager) DeleteRateLimitPolicy(ctx context.Context, filter *rate_limit_policy.RateLimitPolicyFilter) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, filter)
	}
	return nil
}

var _ rate_limit_policy.RateLimitPolicyStorager = (*fakeRateLimitPolicyStorager)(nil)

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
	createFn    func(ctx context.Context, ruleType string, owner *string, param *shared.RouteRulesParam) (int64, error)
	fetchFn     func(ctx context.Context, ruleType string, owner *string) (*shared.RouteRulesParam, error)
	listFn      func(ctx context.Context, filter *shared.RouteRulesFilter) ([]*shared.RouteTableParam, int64, error)
	updateFn    func(ctx context.Context, id int64, param *shared.RouteRulesParam) (int64, error)
	deleteFn    func(ctx context.Context, id int64) error
	fetchByIDFn func(ctx context.Context, id int64) (*shared.RouteRulesParam, error)
	allFn       func(ctx context.Context) ([]*shared.RouteRulesParam, error)

	created     []createRouteRulesCall
	fetched     []fetchRouteRulesCall
	listed      []*shared.RouteRulesFilter
	updated     []updateRouteRulesCall
	deleted     []int64
	fetchedByID []int64
	fetchedAll  int
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

func (s *fakeRouteRulesStorager) FetchAllRouteRules(ctx context.Context) ([]*shared.RouteRulesParam, error) {
	s.fetchedAll++
	if s.allFn != nil {
		return s.allFn(ctx)
	}
	return nil, nil
}

var _ shared.RouteRulesStorager = (*fakeRouteRulesStorager)(nil)

// fakeAPIKeyStorager 实现 api_key.APIKeyStorager
type fakeAPIKeyStorager struct {
	fetchListFn      func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error)
	createFn         func(ctx context.Context, param *api_key.APIKeyParam) (int64, error)
	updateFn         func(ctx context.Context, filter *api_key.APIKeyFilter, param *api_key.APIKeyParam) (int64, error)
	deleteFn         func(ctx context.Context, filter *api_key.APIKeyFilter) error
	createTokenFn    func(ctx context.Context, param *api_key.APIKeyTokenParam) (int64, error)
	updateTokenFn    func(ctx context.Context, filter *api_key.APIKeyTokenFilter, param *api_key.APIKeyTokenParam) error
	fetchTokenListFn func(ctx context.Context, filter *api_key.APIKeyTokenFilter) ([]*api_key.APIKeyTokenParam, error)
}

func (s *fakeAPIKeyStorager) FetchAPIKeyList(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
	if s.fetchListFn != nil {
		return s.fetchListFn(ctx, filter)
	}
	return nil, nil
}

func (s *fakeAPIKeyStorager) CreateAPIKey(ctx context.Context, param *api_key.APIKeyParam) (int64, error) {
	if s.createFn != nil {
		return s.createFn(ctx, param)
	}
	return 0, nil
}

func (s *fakeAPIKeyStorager) UpdateAPIKey(ctx context.Context, filter *api_key.APIKeyFilter, param *api_key.APIKeyParam) (int64, error) {
	if s.updateFn != nil {
		return s.updateFn(ctx, filter, param)
	}
	return 0, nil
}

func (s *fakeAPIKeyStorager) DeleteAPIKey(ctx context.Context, filter *api_key.APIKeyFilter) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, filter)
	}
	return nil
}

func (s *fakeAPIKeyStorager) CreateAPIKeyToken(ctx context.Context, param *api_key.APIKeyTokenParam) (int64, error) {
	if s.createTokenFn != nil {
		return s.createTokenFn(ctx, param)
	}
	return 0, nil
}

func (s *fakeAPIKeyStorager) UpdateAPIKeyToken(ctx context.Context, filter *api_key.APIKeyTokenFilter, param *api_key.APIKeyTokenParam) error {
	if s.updateTokenFn != nil {
		return s.updateTokenFn(ctx, filter, param)
	}
	return nil
}

func (s *fakeAPIKeyStorager) FetchAPIKeyTokenList(ctx context.Context, filter *api_key.APIKeyTokenFilter) ([]*api_key.APIKeyTokenParam, error) {
	if s.fetchTokenListFn != nil {
		return s.fetchTokenListFn(ctx, filter)
	}
	return nil, nil
}

var _ api_key.APIKeyStorager = (*fakeAPIKeyStorager)(nil)

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
