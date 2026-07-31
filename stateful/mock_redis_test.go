// Copyright(c) 2026 Beijing Yingfei Networks Technology Co.Ltd.
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

package stateful

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockRedisClient_IncrAndGet(t *testing.T) {
	client := NewMockRedisClient()

	val, err := client.Incr("counter")
	require.NoError(t, err)
	assert.Equal(t, int64(1), val)

	val, err = client.IncrBy("counter", 4)
	require.NoError(t, err)
	assert.Equal(t, int64(5), val)

	got, err := client.GetInt64("counter")
	require.NoError(t, err)
	assert.Equal(t, int64(5), got)
}

func TestMockRedisClient_Decr(t *testing.T) {
	client := NewMockRedisClient()

	_, err := client.Incr("balance")
	require.NoError(t, err)

	val, err := client.Decr("balance")
	require.NoError(t, err)
	assert.Equal(t, int64(0), val)
}

func TestMockRedisClient_GetMissingKey(t *testing.T) {
	client := NewMockRedisClient()

	v, err := client.Get("missing")
	require.NoError(t, err)
	assert.Nil(t, v)

	got, err := client.GetInt64("missing")
	require.NoError(t, err)
	assert.Equal(t, int64(0), got)
}

func TestMockRedisClient_SetexAndExpire(t *testing.T) {
	client := NewMockRedisClient()

	require.NoError(t, client.Setex("key", []byte("value"), 60))
	require.NoError(t, client.Expire("key", 120))

	got, err := client.GetInt64("key")
	require.NoError(t, err)
	assert.Equal(t, int64(0), got)
}

func TestMockRedisClient_PIncr(t *testing.T) {
	client := NewMockRedisClient()

	results, err := client.PIncr([]string{"a", "b", "a"})
	require.NoError(t, err)
	require.Len(t, results, 3)
	assert.Equal(t, int64(1), results[0])
	assert.Equal(t, int64(1), results[1])
	assert.Equal(t, int64(2), results[2])
}

func TestMockRedisClient_IncrAndExpire(t *testing.T) {
	client := NewMockRedisClient()

	val, err := client.IncrAndExpire("quota", 3600)
	require.NoError(t, err)
	assert.Equal(t, int64(1), val)
}

func TestMockRedisClient_Reset(t *testing.T) {
	client := NewMockRedisClient()

	_, err := client.Incr("x")
	require.NoError(t, err)

	client.Reset()

	got, err := client.GetInt64("x")
	require.NoError(t, err)
	assert.Equal(t, int64(0), got)
}
