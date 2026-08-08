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

package model_provider_type

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iauth"
)

const DefaultModelProviderConfigPath = "conf/ai/models.json"

// modelProviderConfigReader abstracts the config source so tests can inject
// fake readers without touching the file system.
type modelProviderConfigReader func() ([]byte, error)

// DefaultModelProviderConfigReader returns a reader that loads the provider
// type configuration from the default file path.
func DefaultModelProviderConfigReader() modelProviderConfigReader {
	return func() ([]byte, error) {
		return os.ReadFile(DefaultModelProviderConfigPath)
	}
}

// NewEndpoints creates the model-provider-types endpoints using the given
// config reader. Passing nil falls back to the default file-based reader.
func NewEndpoints(reader modelProviderConfigReader) []*xreq.Endpoint {
	if reader == nil {
		reader = DefaultModelProviderConfigReader()
	}

	action := func(req *http.Request) (interface{}, error) {
		return listModelProviderTypesProcess(req.Context(), reader)
	}

	return []*xreq.Endpoint{{
		Path:       "/model-provider-types",
		Method:     http.MethodGet,
		Handler:    xreq.Convert(action),
		Authorizer: iauth.FA(iauth.FeatureProductCluster, iauth.ActionRead),
	}}
}

type ModelProvider struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

func listModelProviderTypesProcess(ctx context.Context, reader modelProviderConfigReader) (interface{}, error) {
	data, err := reader()
	if err != nil {
		return nil, xerror.WrapParamError(fmt.Errorf("failed to read config %s: %w", DefaultModelProviderConfigPath, err))
	}

	var providers []ModelProvider
	if err := json.Unmarshal(data, &providers); err != nil {
		return nil, xerror.WrapParamError(fmt.Errorf("failed to parse config %s: %w", DefaultModelProviderConfigPath, err))
	}

	types := make([]string, 0, len(providers))
	for _, p := range providers {
		if p.ID != "" {
			types = append(types, p.ID)
		}
	}

	return types, nil
}
