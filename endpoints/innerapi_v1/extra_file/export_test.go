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

package extra_file

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/ibasic"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful/container"
)

type fakeExtraFileStorager struct {
	fetchExtraFilesFn func(ctx context.Context, filter *ibasic.ExtraFileFilter) ([]*ibasic.ExtraFile, error)
}

func (f *fakeExtraFileStorager) CreateExtraFile(ctx context.Context, product *ibasic.Product, files ...*ibasic.ExtraFileParam) error {
	return nil
}

func (f *fakeExtraFileStorager) DeleteExtraFile(ctx context.Context, filter *ibasic.ExtraFileFilter) error {
	return nil
}

func (f *fakeExtraFileStorager) FetchExtraFiles(ctx context.Context, filter *ibasic.ExtraFileFilter) ([]*ibasic.ExtraFile, error) {
	if f.fetchExtraFilesFn != nil {
		return f.fetchExtraFilesFn(ctx, filter)
	}
	return nil, nil
}

func setupExtraFileManager(storager ibasic.ExtraFileStorager) func() {
	old := container.ExtraFileManager
	container.ExtraFileManager = ibasic.NewExtraFileManager(storager)
	return func() {
		container.ExtraFileManager = old
	}
}

func TestExportExtraFileActionProcess_Success(t *testing.T) {
	defer setupExtraFileManager(&fakeExtraFileStorager{
		fetchExtraFilesFn: func(ctx context.Context, filter *ibasic.ExtraFileFilter) ([]*ibasic.ExtraFile, error) {
			require.NotNil(t, filter.Name)
			assert.Equal(t, "test.txt", *filter.Name)
			return []*ibasic.ExtraFile{
				{ID: 1, Name: "test.txt", Content: []byte("hello world")},
			}, nil
		},
	})()

	req := httptest.NewRequest(http.MethodGet, "/configs/extra_files/test.txt", nil)
	data, err := ExportExtraFileAction(req)

	require.NoError(t, err)
	require.NotNil(t, data)
	assert.Equal(t, []byte("hello world"), data)
}

func TestExportExtraFileActionProcess_NotFound(t *testing.T) {
	defer setupExtraFileManager(&fakeExtraFileStorager{
		fetchExtraFilesFn: func(ctx context.Context, filter *ibasic.ExtraFileFilter) ([]*ibasic.ExtraFile, error) {
			return nil, nil
		},
	})()

	req := httptest.NewRequest(http.MethodGet, "/configs/extra_files/missing.txt", nil)
	data, err := ExportExtraFileAction(req)

	require.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "Record Not Exist")
}
