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

package testutil

import (
	"context"

	"github.com/rainway-ai-gateway/ai-gateway-api/model/itxn"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iversion_control"
)

// FakeTxn is a no-op transaction storager used by unit tests.
type FakeTxn struct{}

func (f *FakeTxn) AtomExecute(ctx context.Context, do func(context.Context) error) error {
	return do(ctx)
}

var _ itxn.TxnStorager = (*FakeTxn)(nil)

// FakeVersionControlStorager implements iversion_control.VersionControlStorager.
type FakeVersionControlStorager struct {
	Version string
}

func (f *FakeVersionControlStorager) UpsertConfigLastExportedVersion(ctx context.Context, css *iversion_control.ExportData) (string, error) {
	if f.Version != "" {
		return f.Version, nil
	}
	return "v1", nil
}

var _ iversion_control.VersionControlStorager = (*FakeVersionControlStorager)(nil)

// NewVersionControlManager creates a VersionControlManager backed by fake storagers.
func NewVersionControlManager(version string) *iversion_control.VersionControlManager {
	return iversion_control.NewVersionControllerManager(&FakeTxn{}, &FakeVersionControlStorager{Version: version})
}
