// Copyright(c) 2026 Beijing Yingfei Networks Technology Co.Ltd.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http: //www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package quota

import (
	"context"
	"encoding/json"

	"github.com/yf-networks/ai-gateway-api/lib"
	"github.com/yf-networks/ai-gateway-api/model/shared"
	"github.com/yf-networks/ai-gateway-api/storage/rdb/internal/dao"
)

type RouteRulesStorager struct {
	dbCtxFactory lib.DBContextFactory
}

func NewRouteRulesStorager(dbCtxFactory lib.DBContextFactory) *RouteRulesStorager {
	return &RouteRulesStorager{
		dbCtxFactory: dbCtxFactory,
	}
}

var _ shared.RouteRulesStorager = &RouteRulesStorager{}

func (s *RouteRulesStorager) CreateRouteRules(ctx context.Context, ruleType string, owner *string, param *shared.RouteRulesParam) (int64, error) {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return 0, err
	}

	rulesJSON, err := marshalRouteRules(param.Rules)
	if err != nil {
		return 0, err
	}

	data := &dao.TRouteRulesParam{
		Type:    &ruleType,
		Owner:   owner,
		Enabled: param.Enabled,
		Rules:   rulesJSON,
	}

	return dao.TRouteRulesCreate(dbCtx, data)
}

func (s *RouteRulesStorager) FetchRouteRules(ctx context.Context, ruleType string, owner *string) (*shared.RouteRulesParam, error) {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return nil, err
	}

	where := &dao.TRouteRulesParam{
		Type:  &ruleType,
		Owner: owner,
	}
	one, err := dao.TRouteRulesOne(dbCtx, where)
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}

	return routeRulesDataToParam(one), nil
}

func (s *RouteRulesStorager) UpdateRouteRules(ctx context.Context, id int64, param *shared.RouteRulesParam) (int64, error) {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return 0, err
	}

	rulesJSON, err := marshalRouteRules(param.Rules)
	if err != nil {
		return 0, err
	}

	data := &dao.TRouteRulesParam{
		Enabled: param.Enabled,
		Rules:   rulesJSON,
	}

	return dao.TRouteRulesUpdate(dbCtx, data, &dao.TRouteRulesParam{ID: &id})
}

func (s *RouteRulesStorager) DeleteRouteRules(ctx context.Context, id int64) error {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return err
	}

	_, err = dao.TRouteRulesDelete(dbCtx, &dao.TRouteRulesParam{ID: &id})
	return err
}

func (s *RouteRulesStorager) FetchRouteRulesByID(ctx context.Context, id int64) (*shared.RouteRulesParam, error) {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return nil, err
	}

	one, err := dao.TRouteRulesOne(dbCtx, &dao.TRouteRulesParam{ID: &id})
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}

	return routeRulesDataToParam(one), nil
}

func marshalRouteRules(rules []*shared.AiRouteRuleParam) (*string, error) {
	if rules == nil {
		return lib.PString("[]"), nil
	}
	data, err := json.Marshal(rules)
	if err != nil {
		return nil, err
	}
	return lib.PString(string(data)), nil
}

func routeRulesDataToParam(one *dao.TRouteRules) *shared.RouteRulesParam {
	param := &shared.RouteRulesParam{
		ID:      &one.ID,
		Enabled: &one.Enabled,
	}

	if one.Rules != "" {
		var rules []*shared.AiRouteRuleParam
		json.Unmarshal([]byte(one.Rules), &rules)
		param.Rules = rules
	}

	return param
}
