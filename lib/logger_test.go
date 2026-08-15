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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogContext(t *testing.T) {
	ctx := context.Background()
	ctxWithID := NewLogContext(ctx)
	assert.NotNil(t, ctxWithID)
	assert.NotEmpty(t, GainLogID(ctxWithID))

	// Calling again should return the same context/value.
	ctxAgain := NewLogContext(ctxWithID)
	assert.Equal(t, GainLogID(ctxWithID), GainLogID(ctxAgain))
}

func TestGainLogIDMissing(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, "", GainLogID(ctx))
}
