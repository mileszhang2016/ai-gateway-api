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

package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
)

func TestTokenCreateParamValidate(t *testing.T) {
	cases := []struct {
		name    string
		param   *TokenCreateParam
		wantErr bool
	}{
		{
			name:    "valid",
			param:   &TokenCreateParam{Name: lib.PString("token_1"), Scope: lib.PString("System")},
			wantErr: false,
		},
		{
			name:    "invalid token name",
			param:   &TokenCreateParam{Name: lib.PString("-token"), Scope: lib.PString("System")},
			wantErr: true,
		},
		{
			name:    "invalid scope",
			param:   &TokenCreateParam{Name: lib.PString("token_1"), Scope: lib.PString("Other")},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.param.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserCreateParamValidate(t *testing.T) {
	cases := []struct {
		name    string
		param   *UserCreateParam
		wantErr bool
	}{
		{
			name:    "valid",
			param:   &UserCreateParam{UserName: lib.PString("user1"), Password: lib.PString("Password123"), IsAdmin: true},
			wantErr: false,
		},
		{
			name:    "reserved user name",
			param:   &UserCreateParam{UserName: lib.PString("admin"), Password: lib.PString("Password123"), IsAdmin: true},
			wantErr: true,
		},
		{
			name:    "password too short",
			param:   &UserCreateParam{UserName: lib.PString("user1"), Password: lib.PString("short1"), IsAdmin: true},
			wantErr: true,
		},
		{
			name:    "password equals user name",
			param:   &UserCreateParam{UserName: lib.PString("user1"), Password: lib.PString("user1"), IsAdmin: true},
			wantErr: true,
		},
		{
			name:    "is_admin false",
			param:   &UserCreateParam{UserName: lib.PString("user1"), Password: lib.PString("Password123"), IsAdmin: false},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.param.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserUpdateIsAdminParamValidate(t *testing.T) {
	cases := []struct {
		name    string
		param   *UserUpdateIsAdminParam
		wantErr bool
	}{
		{
			name:    "valid",
			param:   &UserUpdateIsAdminParam{UserName: lib.PString("user1"), IsAdmin: true},
			wantErr: false,
		},
		{
			name:    "invalid user name",
			param:   &UserUpdateIsAdminParam{UserName: lib.PString("-user"), IsAdmin: true},
			wantErr: true,
		},
		{
			name:    "is_admin false",
			param:   &UserUpdateIsAdminParam{UserName: lib.PString("user1"), IsAdmin: false},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.param.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserUpdatePasswordParamValidate(t *testing.T) {
	cases := []struct {
		name    string
		param   *UserUpdatePasswordParam
		wantErr bool
	}{
		{
			name:    "valid",
			param:   &UserUpdatePasswordParam{UserName: lib.PString("user1"), Password: lib.PString("NewPass123")},
			wantErr: false,
		},
		{
			name:    "invalid user name",
			param:   &UserUpdatePasswordParam{UserName: lib.PString("admin"), Password: lib.PString("NewPass123")},
			wantErr: true,
		},
		{
			name:    "password too short",
			param:   &UserUpdatePasswordParam{UserName: lib.PString("user1"), Password: lib.PString("short")},
			wantErr: true,
		},
		{
			name:    "password equals user name",
			param:   &UserUpdatePasswordParam{UserName: lib.PString("user1"), Password: lib.PString("user1")},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.param.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
