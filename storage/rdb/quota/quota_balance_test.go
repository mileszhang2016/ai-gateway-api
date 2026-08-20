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
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/rainway-ai-gateway/ai-gateway-api/model/quota"
	"github.com/rainway-ai-gateway/ai-gateway-api/storage/rdb/internal/dao"
)

func TestQuotaBalanceFilterToParam(t *testing.T) {
	id := int64(1)
	planID := int64(2)

	param := quotaBalanceFilterToParam(&quota.QuotaBalanceFilter{
		ID:          &id,
		QuotaPlanID: &planID,
	})

	assert.Equal(t, &id, param.ID)
	assert.Equal(t, &planID, param.QuotaPlanID)

	assert.Nil(t, quotaBalanceFilterToParam(nil))
}

func TestQuotaBalanceDataToParam(t *testing.T) {
	planID := int64(10)
	used := float64(5)
	remaining := float64(95)
	lastReset := time.Now()

	param := quotaBalanceDataToParam(&quota.QuotaBalanceParam{
		QuotaPlanID: &planID,
		Used:        &used,
		Remaining:   &remaining,
		LastResetAt: &lastReset,
	})

	assert.Equal(t, &planID, param.QuotaPlanID)
	assert.NotNil(t, param.Used)
	assert.True(t, decimal.NewFromFloat(used).Equal(*param.Used))
	assert.NotNil(t, param.Remaining)
	assert.True(t, decimal.NewFromFloat(remaining).Equal(*param.Remaining))
	assert.Equal(t, &lastReset, param.LastResetAt)
}

func TestQuotaBalanceParamToData(t *testing.T) {
	now := time.Now()
	data := &dao.TQuotaBalance{
		ID:          1,
		QuotaPlanID: 2,
		Used:        decimal.NewFromInt(3),
		Remaining:   decimal.NewFromInt(4),
		LastResetAt: &now,
	}

	param := quotaBalanceParamToData(data)

	assert.Equal(t, int64(1), *param.ID)
	assert.Equal(t, int64(2), *param.QuotaPlanID)
	assert.Equal(t, float64(3), *param.Used)
	assert.Equal(t, float64(4), *param.Remaining)
	assert.Equal(t, &now, param.LastResetAt)
}
