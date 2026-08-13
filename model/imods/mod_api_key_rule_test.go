// Copyright(c) 2026 The Infinity AI Gateway Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package imods

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iai_route"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iversion_control"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/quota"
)

func TestNewAPIKeyRuleManager(t *testing.T) {
	m := NewAPIKeyRuleManager(
		&fakeTxn{},
		iversion_control.NewVersionControllerManager(&fakeTxn{}, &fakeVersionControlStorager{}),
		&fakeAPIKeyStorager{},
		&fakeAIRouteRuleStorager{},
		&fakeQuotaPlanStorager{},
		&fakeEntityStorager{},
	)
	assert.NotNil(t, m)
	assert.NotNil(t, m.txn)
	assert.NotNil(t, m.versionControlManager)
}

func TestConvertAPIKeyRulesToBfeRules(t *testing.T) {
	rules := []*APIKeyRule{
		{
			Cond: "default_t()",
			Actions: []Action{
				{Cmd: APIKeyActionCMD},
			},
		},
		{
			Cond: "req_host_in(\"example.com\")",
			Actions: []Action{
				{Cmd: APIKeyActionCMD},
				{Cmd: "OTHER"},
			},
		},
	}

	exportRules := convertAPIKeyRulesToBfeRules(rules)
	assert.Len(t, exportRules, 2)
	assert.Equal(t, "default_t()", *exportRules[0].Cond)
	assert.Equal(t, APIKeyActionCMD, exportRules[0].Action.Cmd)
	assert.Equal(t, "req_host_in(\"example.com\")", *exportRules[1].Cond)
	assert.Equal(t, APIKeyActionCMD, exportRules[1].Action.Cmd)
}

func TestConvertAPIKeyRulesToBfeRules_EmptyActions(t *testing.T) {
	rules := []*APIKeyRule{
		{
			Cond:    "default_t()",
			Actions: nil,
		},
	}

	exportRules := convertAPIKeyRulesToBfeRules(rules)
	assert.Len(t, exportRules, 1)
	assert.Equal(t, "default_t()", *exportRules[0].Cond)
	assert.Nil(t, exportRules[0].Action)
}

func TestConvertQuotaPlanToExport(t *testing.T) {
	unlimited := false
	passNoQuota := true
	quotaVal := float64(1000)
	period := "monthly"
	createTime := int64(1234567890)

	qp := convertQuotaPlanToExport(&quota.QuotaPlanParam{
		Unlimited:             &unlimited,
		PassWhenNoEnoughQuota: &passNoQuota,
		Quota:                 &quotaVal,
		ResetPeriod:           &period,
		CreateTime:            &createTime,
	}, "id-1", "redis-1")

	assert.Equal(t, "id-1", qp.Id)
	assert.Equal(t, "QUOTA_redis-1", qp.RedisKey)
	assert.False(t, qp.Unlimited)
	assert.True(t, qp.PassNoQuota)
	assert.Equal(t, int64(1000), qp.Quota)
	assert.Equal(t, 1, qp.ResetMode)
	assert.Equal(t, int64(1234567890), qp.CreateTime)
	assert.Equal(t, int64(-1), qp.ExpiredTime)
}

func TestConvertQuotaPlanToExport_Weekly(t *testing.T) {
	period := "weekly"
	qp := convertQuotaPlanToExport(&quota.QuotaPlanParam{
		ResetPeriod: &period,
	}, "id-1", "redis-1")
	assert.Equal(t, 1, qp.ResetMode)
}

func TestConvertQuotaPlanToExport_OtherPeriod(t *testing.T) {
	period := "daily"
	qp := convertQuotaPlanToExport(&quota.QuotaPlanParam{
		ResetPeriod: &period,
	}, "id-1", "redis-1")
	assert.Equal(t, 0, qp.ResetMode)
}

func TestConvertQuotaPlanToExport_Unlimited(t *testing.T) {
	unlimited := true
	qp := convertQuotaPlanToExport(&quota.QuotaPlanParam{
		Unlimited: &unlimited,
	}, "id-1", "redis-1")
	assert.True(t, qp.Unlimited)
}

func TestContainsStar(t *testing.T) {
	assert.True(t, containsStar([]string{"a", "*", "b"}))
	assert.False(t, containsStar([]string{"a", "b"}))
	assert.False(t, containsStar(nil))
}

func TestIntersectSlices(t *testing.T) {
	assert.Equal(t, []string{"b", "c"}, intersectSlices([]string{"a", "b", "c"}, []string{"b", "c", "d"}))
	assert.Empty(t, intersectSlices([]string{"a"}, []string{"b"}))
	assert.Empty(t, intersectSlices(nil, []string{"a"}))
}

func TestIntersectAllowModels(t *testing.T) {
	assert.Nil(t, intersectAllowModels(nil))
	assert.Equal(t, []string{"a", "b"}, intersectAllowModels([][]string{{"a", "b"}}))
	assert.Equal(t, []string{"b"}, intersectAllowModels([][]string{{"a", "b"}, {"b", "c"}}))
	assert.Empty(t, intersectAllowModels([][]string{{"a"}, {"b"}}))
}

func TestBuildAIRouteAPIKeyRules(t *testing.T) {
	setupState()
	ctx := context.Background()

	aiRouteStore := &fakeAIRouteRuleStorager{
		fetchFn: func(ctx context.Context, filter *iai_route.AIRouteFilter) ([]*iai_route.Rule, error) {
			return []*iai_route.Rule{
				{
					Basic: &iai_route.BasicInfo{
						Domain: strPtr("example.com"),
					},
				},
			}, nil
		},
	}

	m := NewAPIKeyRuleManager(
		&fakeTxn{},
		iversion_control.NewVersionControllerManager(&fakeTxn{}, &fakeVersionControlStorager{}),
		&fakeAPIKeyStorager{},
		aiRouteStore,
		&fakeQuotaPlanStorager{},
		&fakeEntityStorager{},
	)

	product2config, err := m.buildAIRouteAPIKeyRules(ctx)
	assert.NoError(t, err)
	assert.Len(t, product2config, 1)
	assert.Contains(t, product2config, "AI_product")
	assert.Len(t, product2config["AI_product"].Rules, 2)
	assert.Equal(t, `req_host_in("example.com")`, product2config["AI_product"].Rules[0].Cond)
	assert.Equal(t, "default_t()", product2config["AI_product"].Rules[1].Cond)
}

func TestBuildAIRouteAPIKeyRules_Error(t *testing.T) {
	setupState()
	ctx := context.Background()

	aiRouteStore := &fakeAIRouteRuleStorager{
		fetchFn: func(ctx context.Context, filter *iai_route.AIRouteFilter) ([]*iai_route.Rule, error) {
			return nil, assert.AnError
		},
	}

	m := NewAPIKeyRuleManager(
		&fakeTxn{},
		iversion_control.NewVersionControllerManager(&fakeTxn{}, &fakeVersionControlStorager{}),
		&fakeAPIKeyStorager{},
		aiRouteStore,
		&fakeQuotaPlanStorager{},
		&fakeEntityStorager{},
	)

	product2config, err := m.buildAIRouteAPIKeyRules(ctx)
	assert.Error(t, err)
	assert.Nil(t, product2config)
}
