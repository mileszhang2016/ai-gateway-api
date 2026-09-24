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

package ireport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCostFixedPointToAmount(t *testing.T) {
	assert.Equal(t, 0.0, CostFixedPointToAmount(0))
	assert.InDelta(t, 0.000669, CostFixedPointToAmount(66900), 1e-12)
	assert.InDelta(t, 0.1523, CostFixedPointToAmount(15230000), 1e-12)

	// 业务上限 9000 万元对应的定点值（9e15 < 2^53）可被 float64 精确表示，
	// 换算结果为精确的 9e7，无精度损失。
	assert.Equal(t, 9e7, CostFixedPointToAmount(9_000_000_000_000_000))
}
