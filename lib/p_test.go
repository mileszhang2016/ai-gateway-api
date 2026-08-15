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
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPointerHelpers(t *testing.T) {
	assert.Equal(t, int64(1), *PInt64(1))
	assert.Equal(t, float64(1.5), *PFloat64(1.5))
	assert.Equal(t, "x", *PString("x"))
	assert.Equal(t, uint32(2), *PUint32(2))
	assert.Equal(t, int32(3), *PInt32(3))
	assert.Equal(t, int8(4), *PInt8(4))
	assert.Equal(t, int16(5), *PInt16(5))
	assert.Equal(t, 6, *PInt(6))
	assert.Equal(t, true, *PBool(true))

	now := time.Now()
	assert.True(t, PTime(now).Equal(now))

	pt := PTimeNow()
	assert.NotNil(t, pt)
	assert.False(t, pt.IsZero())
}
