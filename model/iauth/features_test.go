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

package iauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAction_Grant(t *testing.T) {
	action := ActionDeny.Grant(ActionRead).Grant(ActionUpdate)
	assert.True(t, action.IsAllowed(ActionRead))
	assert.True(t, action.IsAllowed(ActionUpdate))
	assert.False(t, action.IsAllowed(ActionDelete))
}

func TestAction_Revoke(t *testing.T) {
	action := actionAll.Revoke(ActionDelete)
	assert.True(t, action.IsAllowed(ActionRead))
	assert.False(t, action.IsAllowed(ActionDelete))
}

func TestAction_IsAllowed(t *testing.T) {
	assert.True(t, (ActionRead | ActionUpdate).IsAllowed(ActionRead))
	assert.False(t, ActionRead.IsAllowed(ActionUpdate))
	assert.False(t, ActionDeny.IsAllowed(ActionRead))
}

func TestNewFeatureAuthorization(t *testing.T) {
	auth := NewFeatureAuthorization(FeatureUser, ActionRead)
	require.NotNil(t, auth.FeatureAuthorizer)
	assert.Equal(t, FeatureUser, auth.FeatureAuthorizer.Feature)
	assert.Equal(t, ActionRead, auth.FeatureAuthorizer.Action)
	assert.False(t, auth.ValidateProduct)
}

func TestNewFeatureAuthorizerWithFactoryWithProduct(t *testing.T) {
	auth := NewFeatureAuthorizerWithFactoryWithProduct(FeatureUser, ActionRead)
	require.NotNil(t, auth.FeatureAuthorizer)
	assert.True(t, auth.ValidateProduct)
}
