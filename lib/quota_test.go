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

package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuotaToRedisValue(t *testing.T) {
	rmb := "RMB"
	token := "total_token"

	assert.Equal(t, int64(0), QuotaToRedisValue(nil, &rmb))

	v := float64(1.5)
	assert.Equal(t, int64(150000000), QuotaToRedisValue(&v, &rmb))
	assert.Equal(t, int64(1), QuotaToRedisValue(&v, &token))

	v = float64(0.00000001)
	assert.Equal(t, int64(1), QuotaToRedisValue(&v, &rmb))

	v = float64(100)
	assert.Equal(t, int64(100), QuotaToRedisValue(&v, nil))
}

func TestRedisValueToQuota(t *testing.T) {
	rmb := "RMB"
	token := "total_token"

	assert.Equal(t, float64(1.5), RedisValueToQuota(150000000, &rmb))
	assert.Equal(t, float64(100), RedisValueToQuota(100, &token))
	assert.Equal(t, float64(100), RedisValueToQuota(100, nil))
}
