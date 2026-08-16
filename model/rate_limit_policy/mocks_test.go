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

package rate_limit_policy

import (
	"context"

	"github.com/infinity-ai-gateway/ai-gateway-api/model/api_key"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/entity"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/itxn"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iversion_control"
)

type fakeTxn struct{}

func (f *fakeTxn) AtomExecute(ctx context.Context, do func(context.Context) error) error {
	return do(ctx)
}

var _ itxn.TxnStorager = (*fakeTxn)(nil)

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
