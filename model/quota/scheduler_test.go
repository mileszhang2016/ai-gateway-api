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
	"testing"

	"github.com/stretchr/testify/assert"
)

// spyBalanceSyncer implements BalanceSyncer and records method calls.
type spyBalanceSyncer struct {
	resetExpiredCalled bool
}

func (s *spyBalanceSyncer) ResetExpiredBalances(ctx context.Context) error {
	s.resetExpiredCalled = true
	return nil
}

func TestQuotaResetScheduler_ResetQuotas(t *testing.T) {
	spy := &spyBalanceSyncer{}
	scheduler := NewQuotaResetScheduler(&fakeTxn{}, spy)

	scheduler.resetQuotas()

	assert.True(t, spy.resetExpiredCalled, "ResetExpiredBalances should be called")
}

func TestQuotaResetScheduler_StartStop(t *testing.T) {
	spy := &spyBalanceSyncer{}
	scheduler := NewQuotaResetScheduler(&fakeTxn{}, spy)

	scheduler.Start()
	scheduler.Stop()
}
