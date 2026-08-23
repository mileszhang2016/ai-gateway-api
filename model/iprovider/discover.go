// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package iprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xerror"
)

// DiscoverCaller abstracts the HTTP call used by DiscoverModels.
type DiscoverCaller interface {
	Call(ctx context.Context, method, url string, headers map[string]string) ([]byte, error)
}

// HTTPDiscoverCaller is the default DiscoverCaller using net/http.
type HTTPDiscoverCaller struct {
	Client *http.Client
}

// NewHTTPDiscoverCaller creates a default HTTPDiscoverCaller.
func NewHTTPDiscoverCaller() *HTTPDiscoverCaller {
	return &HTTPDiscoverCaller{
		Client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Call performs an HTTP GET request and returns the response body.
func (c *HTTPDiscoverCaller) Call(ctx context.Context, method, url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("model discovery returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// DiscoverModels triggers model discovery for the named provider and updates its models list.
// It uses the first model protocol to choose the auth header style and the first instance pool
// entry as the request target. The discovered models are written back to the provider record.
func (m *ProviderManager) DiscoverModels(ctx context.Context, name string) ([]string, error) {
	return m.DiscoverModelsWithCaller(ctx, name, NewHTTPDiscoverCaller())
}

// DiscoverModelsWithCaller triggers model discovery using the supplied caller (for tests).
func (m *ProviderManager) DiscoverModelsWithCaller(ctx context.Context, name string, caller DiscoverCaller) ([]string, error) {
	var models []string
	err := m.txn.AtomExecute(ctx, func(ctx context.Context) error {
		provider, err := m.storager.FetchProvider(ctx, &ProviderFilter{Name: &name})
		if err != nil {
			return err
		}
		if provider == nil {
			return xerror.WrapRecordNotExist("provider")
		}

		if len(provider.Keys) == 0 {
			return xerror.WrapParamErrorWithMsg("provider has no keys for authentication")
		}
		if len(provider.InstancePool) == 0 {
			return xerror.WrapParamErrorWithMsg("provider has no instances for discovery")
		}
		if len(provider.ModelProtocols) == 0 {
			return xerror.WrapParamErrorWithMsg("provider has no model_protocols")
		}

		protocol := provider.ModelProtocols[0]
		key := provider.Keys[0].Key
		host := provider.InstancePool[0]
		url := BuildDiscoverURL(provider.ModelEndpoint, host)

		headerName, headerValue := BuildAuthHeader(protocol, key)
		headers := map[string]string{headerName: headerValue}

		body, err := caller.Call(ctx, http.MethodGet, url, headers)
		if err != nil {
			return xerror.WrapParamErrorWithMsg("model discovery request failed: %v", err)
		}

		models, err = ParseModelDiscoveryResponse(body, protocol)
		if err != nil {
			return xerror.WrapParamErrorWithMsg("model discovery response parse failed: %v", err)
		}

		param := &ProviderParam{
			Models: models,
		}
		return m.storager.UpdateProvider(ctx, name, param)
	})
	return models, err
}

// modelParser describes how to extract model names from a provider discovery response.
type modelParser struct {
	ListPath  string // JSON key of the array containing model entries
	IDField   string // field used as the canonical model identifier
	NameField string // fallback field used when IDField is absent or empty
}

// modelProtocolParsers maps model protocol names to their discovery response parsers.
// The definitions formerly lived in conf/ai/model_definition.json; they are now
// maintained in code so that /providers/{name}/discover-models can select the right
// parser based on provider.model_protocols.
var modelProtocolParsers = map[string]modelParser{
	"openai": {
		ListPath:  "data",
		IDField:   "id",
		NameField: "object",
	},
	"anthropic": {
		ListPath:  "models",
		IDField:   "model_id",
		NameField: "display_name",
	},
}

// ParseModelDiscoveryResponse extracts a model name list from a provider discovery response.
func ParseModelDiscoveryResponse(body []byte, protocol string) ([]string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	// Use protocol-specific parser when available.
	if parser, ok := modelProtocolParsers[protocol]; ok {
		return parseWithModelParser(data, parser, protocol)
	}

	// Fallback to generic discovery heuristics for unknown protocols.
	return parseModelDiscoveryResponseGeneric(data, protocol)
}

func parseWithModelParser(data map[string]interface{}, parser modelParser, protocol string) ([]string, error) {
	list, ok := data[parser.ListPath].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unable to extract models from %s response: %q array not found", protocol, parser.ListPath)
	}

	models := make([]string, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if id, ok := m[parser.IDField].(string); ok && id != "" {
			models = append(models, id)
			continue
		}
		if parser.NameField != "" {
			if name, ok := m[parser.NameField].(string); ok && name != "" {
				models = append(models, name)
			}
		}
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("unable to extract models from %s response: no valid entries in %q array", protocol, parser.ListPath)
	}
	return models, nil
}

func parseModelDiscoveryResponseGeneric(data map[string]interface{}, protocol string) ([]string, error) {
	// Try common OpenAI-style "data" array first.
	if list, ok := data["data"].([]interface{}); ok {
		models := make([]string, 0, len(list))
		for _, item := range list {
			if m, ok := item.(map[string]interface{}); ok {
				if id, ok := m["id"].(string); ok && id != "" {
					models = append(models, id)
					continue
				}
				if name, ok := m["name"].(string); ok && name != "" {
					models = append(models, name)
				}
			}
		}
		if len(models) > 0 {
			return models, nil
		}
	}

	// Fallback to "models" array.
	if list, ok := data["models"].([]interface{}); ok {
		models := make([]string, 0, len(list))
		for _, item := range list {
			if m, ok := item.(map[string]interface{}); ok {
				if name, ok := m["name"].(string); ok && name != "" {
					models = append(models, name)
					continue
				}
				if id, ok := m["id"].(string); ok && id != "" {
					models = append(models, id)
				}
			} else if s, ok := item.(string); ok && s != "" {
				models = append(models, s)
			}
		}
		if len(models) > 0 {
			return models, nil
		}
	}

	// Some providers return { "models": { "name": ... } } for a single model.
	if single, ok := data["models"].(map[string]interface{}); ok {
		if name, ok := single["name"].(string); ok && name != "" {
			return []string{name}, nil
		}
	}

	return nil, fmt.Errorf("unable to extract models from %s response", protocol)
}
