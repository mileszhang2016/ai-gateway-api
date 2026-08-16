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

package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/ibasic"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful/container"
)

type fakeAuthenticateStorager struct {
	fetchUserListFn func(ctx context.Context, param *iauth.UserFilter) ([]*iauth.User, error)
	fetchUserFn     func(ctx context.Context, param *iauth.UserFilter) (*iauth.User, error)
	updateUserFn    func(ctx context.Context, user *iauth.User, param *iauth.UserParam) error
	createUserFn    func(ctx context.Context, param *iauth.UserParam) error
	deleteUserFn    func(ctx context.Context, user *iauth.User) error
	fetchTokensFn   func(ctx context.Context, param *iauth.TokenFilter) ([]*iauth.Token, error)
	createTokenFn   func(ctx context.Context, token *iauth.TokenParam) error
	deleteTokenFn   func(ctx context.Context, param *iauth.Token) error
}

func (s *fakeAuthenticateStorager) FetchUserList(ctx context.Context, param *iauth.UserFilter) ([]*iauth.User, error) {
	if s.fetchUserListFn != nil {
		return s.fetchUserListFn(ctx, param)
	}
	return nil, nil
}

func (s *fakeAuthenticateStorager) FetchUser(ctx context.Context, param *iauth.UserFilter) (*iauth.User, error) {
	if s.fetchUserFn != nil {
		return s.fetchUserFn(ctx, param)
	}
	return nil, nil
}

func (s *fakeAuthenticateStorager) UpdateUser(ctx context.Context, user *iauth.User, param *iauth.UserParam) error {
	if s.updateUserFn != nil {
		return s.updateUserFn(ctx, user, param)
	}
	return nil
}

func (s *fakeAuthenticateStorager) CreateUser(ctx context.Context, param *iauth.UserParam) error {
	if s.createUserFn != nil {
		return s.createUserFn(ctx, param)
	}
	return nil
}

func (s *fakeAuthenticateStorager) DeleteUser(ctx context.Context, user *iauth.User) error {
	if s.deleteUserFn != nil {
		return s.deleteUserFn(ctx, user)
	}
	return nil
}

func (s *fakeAuthenticateStorager) FetchTokens(ctx context.Context, param *iauth.TokenFilter) ([]*iauth.Token, error) {
	if s.fetchTokensFn != nil {
		return s.fetchTokensFn(ctx, param)
	}
	return nil, nil
}

func (s *fakeAuthenticateStorager) CreateToken(ctx context.Context, token *iauth.TokenParam) error {
	if s.createTokenFn != nil {
		return s.createTokenFn(ctx, token)
	}
	return nil
}

func (s *fakeAuthenticateStorager) DeleteToken(ctx context.Context, param *iauth.Token) error {
	if s.deleteTokenFn != nil {
		return s.deleteTokenFn(ctx, param)
	}
	return nil
}

type fakeAuthorizeStorager struct{}

func (s *fakeAuthorizeStorager) UnbindUserProduct(ctx context.Context, user *iauth.User, product *ibasic.Product) error {
	return nil
}

func (s *fakeAuthorizeStorager) UnbindUserAllProduct(ctx context.Context, user *iauth.User) error {
	return nil
}

func (s *fakeAuthorizeStorager) BindUserProduct(ctx context.Context, user *iauth.User, product *ibasic.Product) error {
	return nil
}

func (s *fakeAuthorizeStorager) FetchUserProducts(ctx context.Context, user *iauth.User) ([]*ibasic.Product, error) {
	return nil, nil
}

func (s *fakeAuthorizeStorager) FetchProductUsers(ctx context.Context, product *ibasic.Product) ([]*iauth.User, error) {
	return nil, nil
}

func (s *fakeAuthorizeStorager) UpdateUserScopes(ctx context.Context, user *iauth.User, scopes []string) error {
	return nil
}

func (s *fakeAuthorizeStorager) IsUserProductGranted(ctx context.Context, user *iauth.User, product *ibasic.Product) (bool, error) {
	return false, nil
}

func (s *fakeAuthorizeStorager) UnbindTokenAllProduct(ctx context.Context, token *iauth.Token) error {
	return nil
}

func (s *fakeAuthorizeStorager) BindTokenProduct(ctx context.Context, token *iauth.Token, product *ibasic.Product) error {
	return nil
}

func (s *fakeAuthorizeStorager) FetchProductTokens(ctx context.Context, product *ibasic.Product) ([]*iauth.Token, error) {
	return nil, nil
}

func (s *fakeAuthorizeStorager) IsTokenProductGranted(ctx context.Context, token *iauth.Token, product *ibasic.Product) (bool, error) {
	return false, nil
}

func (s *fakeAuthorizeStorager) FetchTokenProduct(ctx context.Context, token *iauth.Token) (*ibasic.Product, error) {
	return nil, nil
}

func (s *fakeAuthorizeStorager) BatchFetchTokenProduct(ctx context.Context, token []*iauth.Token) (map[int64]*ibasic.Product, error) {
	return nil, nil
}

func setAuthenticateManager(storager iauth.AuthenticateStorager) func() {
	old := container.AuthenticateManager
	container.AuthenticateManager = iauth.NewAuthenticateManager(&fakeTxn{}, storager, &fakeAuthorizeStorager{})
	return func() {
		container.AuthenticateManager = old
	}
}


func TestUserProbeAction_NoAuthorization(t *testing.T) {
	defer setAuthenticateManager(&fakeAuthenticateStorager{})()

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	newReq, err := UserProbeAction(req)

	require.NoError(t, err)
	assert.Equal(t, req, newReq)
}

func TestUserProbeAction_BadFormat(t *testing.T) {
	defer setAuthenticateManager(&fakeAuthenticateStorager{})()

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	req.Header.Set("Authorization", "OnlyOnePart")
	newReq, err := UserProbeAction(req)

	require.Error(t, err)
	assert.Nil(t, newReq)
	assert.Contains(t, err.Error(), "Bad Format Header Authorization")
}

func TestUserProbeAction_AuthenticateFail(t *testing.T) {
	defer setAuthenticateManager(&fakeAuthenticateStorager{
		fetchTokensFn: func(ctx context.Context, param *iauth.TokenFilter) ([]*iauth.Token, error) {
			return nil, errors.New("auth failed")
		},
	})()

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	req.Header.Set("Authorization", "Token fake-token")
	newReq, err := UserProbeAction(req)

	require.Error(t, err)
	assert.Nil(t, newReq)
	assert.Equal(t, "auth failed", err.Error())
}

func TestUserProbeAction_Success(t *testing.T) {
	expected := &iauth.Visitor{
		Token: &iauth.Token{ID: 1, Name: "token-user", Token: "fake-token"},
	}
	defer setAuthenticateManager(&fakeAuthenticateStorager{
		fetchTokensFn: func(ctx context.Context, param *iauth.TokenFilter) ([]*iauth.Token, error) {
			require.NotNil(t, param.Token)
			assert.Equal(t, "fake-token", *param.Token)
			return []*iauth.Token{expected.Token}, nil
		},
	})()

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	req.Header.Set("Authorization", "Token fake-token")
	newReq, err := UserProbeAction(req)

	require.NoError(t, err)
	require.NotNil(t, newReq)

	visitor, err := iauth.MustGetVisitor(newReq.Context())
	require.NoError(t, err)
	assert.Equal(t, expected.Token, visitor.Token)
}
