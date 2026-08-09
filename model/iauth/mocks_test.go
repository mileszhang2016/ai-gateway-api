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
	"context"

	"github.com/infinity-ai-gateway/ai-gateway-api/model/ibasic"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/itxn"
)

// fakeTxn implements itxn.TxnStorager for unit tests
type fakeTxn struct{}

func (f *fakeTxn) AtomExecute(ctx context.Context, do func(context.Context) error) error {
	return do(ctx)
}

var _ itxn.TxnStorager = (*fakeTxn)(nil)

// fakeAuthenticateStorager implements AuthenticateStorager
type fakeAuthenticateStorager struct {
	fetchUserListFn func(ctx context.Context, param *UserFilter) ([]*User, error)
	fetchUserFn     func(ctx context.Context, param *UserFilter) (*User, error)
	updateUserFn    func(ctx context.Context, user *User, param *UserParam) error
	createUserFn    func(ctx context.Context, param *UserParam) error
	deleteUserFn    func(ctx context.Context, user *User) error
	fetchTokensFn   func(ctx context.Context, param *TokenFilter) ([]*Token, error)
	createTokenFn   func(ctx context.Context, token *TokenParam) error
	deleteTokenFn   func(ctx context.Context, param *Token) error
}

func (s *fakeAuthenticateStorager) FetchUserList(ctx context.Context, param *UserFilter) ([]*User, error) {
	if s.fetchUserListFn != nil {
		return s.fetchUserListFn(ctx, param)
	}
	return nil, nil
}

func (s *fakeAuthenticateStorager) FetchUser(ctx context.Context, param *UserFilter) (*User, error) {
	if s.fetchUserFn != nil {
		return s.fetchUserFn(ctx, param)
	}
	return nil, nil
}

func (s *fakeAuthenticateStorager) UpdateUser(ctx context.Context, user *User, param *UserParam) error {
	if s.updateUserFn != nil {
		return s.updateUserFn(ctx, user, param)
	}
	return nil
}

func (s *fakeAuthenticateStorager) CreateUser(ctx context.Context, param *UserParam) error {
	if s.createUserFn != nil {
		return s.createUserFn(ctx, param)
	}
	return nil
}

func (s *fakeAuthenticateStorager) DeleteUser(ctx context.Context, user *User) error {
	if s.deleteUserFn != nil {
		return s.deleteUserFn(ctx, user)
	}
	return nil
}

func (s *fakeAuthenticateStorager) FetchTokens(ctx context.Context, param *TokenFilter) ([]*Token, error) {
	if s.fetchTokensFn != nil {
		return s.fetchTokensFn(ctx, param)
	}
	return nil, nil
}

func (s *fakeAuthenticateStorager) CreateToken(ctx context.Context, token *TokenParam) error {
	if s.createTokenFn != nil {
		return s.createTokenFn(ctx, token)
	}
	return nil
}

func (s *fakeAuthenticateStorager) DeleteToken(ctx context.Context, param *Token) error {
	if s.deleteTokenFn != nil {
		return s.deleteTokenFn(ctx, param)
	}
	return nil
}

var _ AuthenticateStorager = (*fakeAuthenticateStorager)(nil)

// fakeAuthorizeStorager implements AuthorizeStorager
type fakeAuthorizeStorager struct {
	unbindUserProductFn      func(ctx context.Context, user *User, product *ibasic.Product) error
	unbindUserAllProductFn   func(ctx context.Context, user *User) error
	bindUserProductFn        func(ctx context.Context, user *User, product *ibasic.Product) error
	fetchUserProductsFn      func(ctx context.Context, user *User) ([]*ibasic.Product, error)
	fetchProductUsersFn      func(ctx context.Context, product *ibasic.Product) ([]*User, error)
	updateUserScopesFn       func(ctx context.Context, user *User, scopes []string) error
	isUserProductGrantedFn   func(ctx context.Context, user *User, product *ibasic.Product) (bool, error)
	unbindTokenAllProductFn  func(ctx context.Context, token *Token) error
	bindTokenProductFn       func(ctx context.Context, token *Token, product *ibasic.Product) error
	fetchProductTokensFn     func(ctx context.Context, product *ibasic.Product) ([]*Token, error)
	isTokenProductGrantedFn  func(ctx context.Context, token *Token, product *ibasic.Product) (bool, error)
	fetchTokenProductFn      func(ctx context.Context, token *Token) (*ibasic.Product, error)
	batchFetchTokenProductFn func(ctx context.Context, token []*Token) (map[int64]*ibasic.Product, error)
}

func (s *fakeAuthorizeStorager) UnbindUserProduct(ctx context.Context, user *User, product *ibasic.Product) error {
	if s.unbindUserProductFn != nil {
		return s.unbindUserProductFn(ctx, user, product)
	}
	return nil
}

func (s *fakeAuthorizeStorager) UnbindUserAllProduct(ctx context.Context, user *User) error {
	if s.unbindUserAllProductFn != nil {
		return s.unbindUserAllProductFn(ctx, user)
	}
	return nil
}

func (s *fakeAuthorizeStorager) BindUserProduct(ctx context.Context, user *User, product *ibasic.Product) error {
	if s.bindUserProductFn != nil {
		return s.bindUserProductFn(ctx, user, product)
	}
	return nil
}

func (s *fakeAuthorizeStorager) FetchUserProducts(ctx context.Context, user *User) ([]*ibasic.Product, error) {
	if s.fetchUserProductsFn != nil {
		return s.fetchUserProductsFn(ctx, user)
	}
	return nil, nil
}

func (s *fakeAuthorizeStorager) FetchProductUsers(ctx context.Context, product *ibasic.Product) ([]*User, error) {
	if s.fetchProductUsersFn != nil {
		return s.fetchProductUsersFn(ctx, product)
	}
	return nil, nil
}

func (s *fakeAuthorizeStorager) UpdateUserScopes(ctx context.Context, user *User, scopes []string) error {
	if s.updateUserScopesFn != nil {
		return s.updateUserScopesFn(ctx, user, scopes)
	}
	return nil
}

func (s *fakeAuthorizeStorager) IsUserProductGranted(ctx context.Context, user *User, product *ibasic.Product) (bool, error) {
	if s.isUserProductGrantedFn != nil {
		return s.isUserProductGrantedFn(ctx, user, product)
	}
	return false, nil
}

func (s *fakeAuthorizeStorager) UnbindTokenAllProduct(ctx context.Context, token *Token) error {
	if s.unbindTokenAllProductFn != nil {
		return s.unbindTokenAllProductFn(ctx, token)
	}
	return nil
}

func (s *fakeAuthorizeStorager) BindTokenProduct(ctx context.Context, token *Token, product *ibasic.Product) error {
	if s.bindTokenProductFn != nil {
		return s.bindTokenProductFn(ctx, token, product)
	}
	return nil
}

func (s *fakeAuthorizeStorager) FetchProductTokens(ctx context.Context, product *ibasic.Product) ([]*Token, error) {
	if s.fetchProductTokensFn != nil {
		return s.fetchProductTokensFn(ctx, product)
	}
	return nil, nil
}

func (s *fakeAuthorizeStorager) IsTokenProductGranted(ctx context.Context, token *Token, product *ibasic.Product) (bool, error) {
	if s.isTokenProductGrantedFn != nil {
		return s.isTokenProductGrantedFn(ctx, token, product)
	}
	return false, nil
}

func (s *fakeAuthorizeStorager) FetchTokenProduct(ctx context.Context, token *Token) (*ibasic.Product, error) {
	if s.fetchTokenProductFn != nil {
		return s.fetchTokenProductFn(ctx, token)
	}
	return nil, nil
}

func (s *fakeAuthorizeStorager) BatchFetchTokenProduct(ctx context.Context, token []*Token) (map[int64]*ibasic.Product, error) {
	if s.batchFetchTokenProductFn != nil {
		return s.batchFetchTokenProductFn(ctx, token)
	}
	return nil, nil
}

var _ AuthorizeStorager = (*fakeAuthorizeStorager)(nil)
