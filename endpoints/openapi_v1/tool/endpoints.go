// Copyright(c) 2026 The Infinity AI Gateway Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http: //www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful"
)

var _ xreq.Handler = GetModelsFromProviderAction

var GetModelsFromProviderRoute = &xreq.Endpoint{
	Path:       "/tools/get-models-from-provider",
	Method:     http.MethodPost,
	Handler:    xreq.Convert(GetModelsFromProviderAction),
	Authorizer: iauth.FA(iauth.FeatureProductCluster, iauth.ActionCreate),
}

var Endpoints = []*xreq.Endpoint{
	GetModelsFromProviderRoute,
}

const (
	timeOutSecond    = 10
	sleepMicroSecond = 10
	retry            = 1
)

type GetModelsFromProviderParam struct {
	Schema       string            `json:"schema" validate:"required,oneof=http https"`
	URI          string            `json:"uri"`
	Hosts        []string          `json:"hosts" validate:"required,min=1"`
	Headers      map[string]string `json:"headers"`
	ProviderType string            `json:"provider_type"`
}

func GetModelsFromProviderAction(req *http.Request) (interface{}, error) {
	param := &GetModelsFromProviderParam{}
	if err := xreq.BindJSON(req, param); err != nil {
		return nil, err
	}

	return getModelsFromProviderProcess(req.Context(), param)
}

func getModelsFromProviderProcess(ctx context.Context, param *GetModelsFromProviderParam) (interface{}, error) {
	staticProviders := map[string]bool{
		"huoshancodeplan": true,
		"bailiantokenplan": true,
	}
	if staticProviders[param.ProviderType] {
		return loadStaticModelList(param.ProviderType)
	}

	parserConf, err := LoadParserConfig("conf/ai/model_definition.json")
	if err != nil {
		return nil, xerror.WrapParamError(err)
	}

	response, err := callModelAPI(ctx, param)
	if err != nil {
		return nil, xerror.WrapParamError(err)
	}

	return ParseModelsWithConfig([]byte(response["result"]), param.ProviderType, parserConf)
}

func loadStaticModelList(providerType string) ([]map[string]interface{}, error) {
	data, err := os.ReadFile("conf/ai/model_list.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read model list config: %w", err)
	}

	var config map[string][]map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse model list config: %w", err)
	}

	if models, ok := config[providerType]; ok {
		return models, nil
	}

	return nil, xerror.WrapParamError(fmt.Errorf("no static models found for provider_type: %s", providerType))
}

func callModelAPI(ctx context.Context, param *GetModelsFromProviderParam) (map[string]string, error) {
	url := param.Schema
	var errStr string

	uri := param.URI
	if uri == "" {
		uri = "/v1/models"
	}

	for i, host := range param.Hosts {
		reqURL := host
		if uri == "/" {
			reqURL = reqURL + uri
		} else {
			if uri != "" {
				reqURL = path.Join(reqURL, uri)
			}
		}
		reqURL = fmt.Sprintf("%s://%s", url, reqURL)

		response, err := lib.ReadWithRetry(reqURL, timeOutSecond, param.Headers, retry, sleepMicroSecond, nil)
		if err != nil {
			eStr := fmt.Sprintf("exec-api, request index:%d url:%s is error:%s", i, reqURL, err.Error())
			stateful.AccessLogger.Error(eStr)
			if errStr == "" {
				errStr = eStr
			} else {
				errStr += "\n" + eStr
			}
			continue
		}

		resp := make(map[string]string)
		resp["result"] = string(response)
		return resp, nil
	}

	return nil, xerror.WrapParamError(fmt.Errorf(errStr))
}

// FieldMapping describes how to map provider JSON into model fields.
type FieldMapping struct {
	ListPath   string            `json:"list_path"`
	IDField    string            `json:"id_field"`
	NameField  string            `json:"name_field"`
	Created    string            `json:"created_field"`
	OwnerField string            `json:"owner_field"`
	Type       string            `json:"type"` // "array", "object", "auto_detect"
	Custom     map[string]string `json:"custom_fields"`
}

type ParserConfig struct {
	Providers     map[string]FieldMapping `json:"providers"`
	DefaultParser FieldMapping            `json:"default_parser"`
}

// LoadParserConfig reads parser config from a JSON file.
func LoadParserConfig(filePath string) (*ParserConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config ParserConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// ParseModelsWithConfig extracts model records from a provider response using config.
func ParseModelsWithConfig(response []byte, provider string, config *ParserConfig) ([]map[string]interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(response, &data); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	// Resolve field mapping for this provider
	parserConfig, exists := config.Providers[provider]
	if !exists {
		parserConfig = config.DefaultParser
	}

	// Extract model list from JSON
	models, err := extractModelList(data, parserConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to extract models: %w", err)
	}

	return models, nil
}

func extractModelList(data map[string]interface{}, config FieldMapping) ([]map[string]interface{}, error) {
	// If list_path is set, drill down using dot-separated segments
	var modelList []interface{}

	if config.ListPath != "" {
		pathParts := strings.Split(config.ListPath, ".")
		current := interface{}(data)

		for _, part := range pathParts {
			if m, ok := current.(map[string]interface{}); ok {
				current = m[part]
			} else if a, ok := current.([]interface{}); ok && part == "*" {
				// Wildcard segment: entire array becomes the model list
				modelList = a
				break
			} else {
				return nil, fmt.Errorf("path %s not found", config.ListPath)
			}
		}

		if current != nil {
			if list, ok := current.([]interface{}); ok {
				modelList = list
			} else {
				return nil, fmt.Errorf("list path does not point to an array")
			}
		}
	}

	// Map each list item through field mapping
	var models []map[string]interface{}
	for _, item := range modelList {
		modelMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		extracted := extractModelFields(modelMap, config)
		if extracted != nil {
			models = append(models, extracted)
		}
	}

	return models, nil
}

func extractModelFields(model map[string]interface{}, config FieldMapping) map[string]interface{} {
	result := make(map[string]interface{})

	// Standard fields from config
	if config.IDField != "" {
		if value, ok := getNestedField(model, config.IDField); ok {
			result["id"] = value
		}
	}

	if config.NameField != "" {
		if value, ok := getNestedField(model, config.NameField); ok {
			result["name"] = value
		}
	}

	if config.Created != "" {
		if value, ok := getNestedField(model, config.Created); ok {
			result["created"] = value
		}
	}

	if config.OwnerField != "" {
		if value, ok := getNestedField(model, config.OwnerField); ok {
			result["owner"] = value
		}
	}

	// Custom fields from config
	for key, path := range config.Custom {
		if value, ok := getNestedField(model, path); ok {
			result[key] = value
		}
	}

	// Skip entries without an id
	if _, hasID := result["id"]; !hasID {
		return nil
	}

	return result
}

// getNestedField reads a dot-separated path from a nested map.
func getNestedField(data map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil, false
		}
	}

	return current, current != nil
}
