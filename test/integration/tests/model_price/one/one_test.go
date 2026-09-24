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

package model_price_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
	"github.com/stretchr/testify/assert"
)

var sm *testutil.ServerManager

func TestMain(m *testing.M) {
	var err error
	sm, err = testutil.StartServer()
	if err != nil {
		panic("failed to start server: " + err.Error())
	}
	code := m.Run()
	sm.Shutdown()
	os.Exit(code)
}

func TestModelPrice_One(t *testing.T) {
	provider := testutil.UniqueName("provider")
	model := "deepseek-v3"

	id, err := testutil.CreateModelPrice(map[string]interface{}{
		"provider":   provider,
		"model":      model,
		"base_model": model,
		"mode":       "chat",
		"prices": map[string]interface{}{
			"input_cost_per_token": 0.000002,
		},
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer testutil.DeleteModelPrice(id)

	t.Run("MP-4-001 按 id 查询存在的记录", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/model-prices/" + fmt.Sprintf("%d", id))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "id", float64(id))
		testutil.AssertDataFieldEquals(t, resp, "provider", provider)
		testutil.AssertDataFieldEquals(t, resp, "model", model)
	})

	t.Run("MP-4-002 按 id 查询不存在的记录", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/model-prices/999999999")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 404)
	})

	t.Run("MP-5-001 按组合键查询存在的记录", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/model-prices", map[string]string{
			"provider": provider,
			"model":    model,
			"mode":     "chat",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		// §3.6 单记录契约：Data 为单个 ModelPrice 对象而非列表包装（issue #170）
		var one map[string]interface{}
		if err := json.Unmarshal(resp.Data, &one); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		assert.NotContains(t, one, "list")
		assert.NotContains(t, one, "pagination")
		assert.Equal(t, float64(id), one["id"])
		assert.Equal(t, provider, one["provider"])
		assert.Equal(t, model, one["model"])
		assert.Equal(t, "chat", one["mode"])
	})

	t.Run("MP-5-002 按组合键查询缺少参数", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/model-prices", map[string]string{
			"provider": provider,
			"model":    model,
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		// 带 model 但三参不齐：参数错误拒绝，不回落列表（issue #170）
		testutil.AssertErrCode(t, resp, 422)
	})

	t.Run("MP-5-003 按组合键查询不存在的记录", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/model-prices", map[string]string{
			"provider": provider,
			"model":    "not-exist",
			"mode":     "chat",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 404)
	})
}
