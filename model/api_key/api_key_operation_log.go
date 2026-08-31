// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package api_key

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rainway-ai-gateway/ai-gateway-api/model/ioperlog"
)

func (rppm *APIKeyManager) recordAPIKeyOperation(ctx context.Context, action string, apiKey *APIKeyParam, before, after map[string]interface{}) {
	if rppm.operationLogManager == nil || apiKey == nil {
		return
	}

	resourceID := ""
	if apiKey.ID != nil {
		resourceID = *apiKey.ID
	}

	resourceName := resourceID
	if apiKey.Description != nil && *apiKey.Description != "" {
		resourceName = *apiKey.Description
	}

	resourceParentID := ""
	if apiKey.EntityID != nil {
		resourceParentID = *apiKey.EntityID
	}

	entry := &ioperlog.OperationLogEntry{
		Action:           action,
		ResourceType:     string(ioperlog.ResourceTypeAPIKey),
		ResourceID:       resourceID,
		ResourceName:     resourceName,
		ResourceParentID: resourceParentID,
		Status:           ioperlog.StatusSuccess,
		CreatedAt:        time.Now(),
	}

	changeSummary := map[string]interface{}{}
	if len(before) > 0 {
		changeSummary["before"] = ioperlog.MaskSensitiveFields(before)
	}
	if len(after) > 0 {
		changeSummary["after"] = ioperlog.MaskSensitiveFields(after)
	}
	if len(changeSummary) > 0 {
		entry.ChangeSummary = changeSummary
	}

	rppm.operationLogManager.Record(ctx, entry)
}

func apiKeyParamToMap(param *APIKeyParam) map[string]interface{} {
	if param == nil {
		return nil
	}

	data, err := json.Marshal(param)
	if err != nil {
		return nil
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}

	return m
}
