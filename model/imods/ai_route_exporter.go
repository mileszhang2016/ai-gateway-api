// Copyright(c) 2026 Beijing Yingfei Networks Technology Co.Ltd.
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

package imods

import (
	"context"
	"fmt"

	"github.com/yf-networks/ai-gateway-api/model/icluster_conf"
	"github.com/yf-networks/ai-gateway-api/model/iversion_control"
	"github.com/yf-networks/ai-gateway-api/model/quota"
	"github.com/yf-networks/ai-gateway-api/model/shared"
)

const ConfigTopicProductAIRoute = "ai_route"

// AiRouteDataExport defines the AI route configuration export structure
type AiRouteDataExport struct {
	Version                  string                         `json:"Version"`
	RouteRules               map[string]*RouteTableExport   `json:"RouteRules"`
	ApikeyRouteTableBindings map[string][]string            `json:"ApikeyRouteTableBindings"`
}

// RouteTableExport defines a single route table
type RouteTableExport struct {
	Type   string             `json:"type"`
	Owner  string             `json:"owner"`
	Rules  []*RouteRuleExport `json:"rules"`
}

// RouteRuleExport defines a route rule
type RouteRuleExport struct {
	Name      string                  `json:"name"`
	Cond      string                  `json:"Cond"`
	Targets   []*AiRouteTargetExport  `json:"targets"`
	Fallbacks []*AiRouteFallbackExport `json:"fallbacks"`
}

// AiRouteTargetExport defines a forwarding target
type AiRouteTargetExport struct {
	ClusterName string `json:"ClusterName"`
	Model       string `json:"Model"`
	Weight      int    `json:"Weight"`
}

// AiRouteFallbackExport defines a fallback target
type AiRouteFallbackExport struct {
	ClusterName string `json:"ClusterName"`
	Model       string `json:"Model"`
}

// UpdateVersion updates the configuration version
func (conf *AiRouteDataExport) UpdateVersion(version string) error {
	conf.Version = version
	return nil
}

// AIRouteExporter manages AI route configuration export
type AIRouteExporter struct {
	apiKeyStorager   icluster_conf.APIKeyStorager
	entityStorager   quota.EntityStorager
	routeRulesStorager shared.RouteRulesStorager
	versionControlManager *iversion_control.VersionControlManager
}

// NewAIRouteExporter creates a new AIRouteExporter instance
func NewAIRouteExporter(apiKeyStorager icluster_conf.APIKeyStorager,
	entityStorager quota.EntityStorager, routeRulesStorager shared.RouteRulesStorager,
	versionControlManager *iversion_control.VersionControlManager) *AIRouteExporter {
	return &AIRouteExporter{
		apiKeyStorager:        apiKeyStorager,
		entityStorager:        entityStorager,
		routeRulesStorager:    routeRulesStorager,
		versionControlManager: versionControlManager,
	}
}

// ConfigExport exports AI route configuration for BFE
func (e *AIRouteExporter) ConfigExport(ctx context.Context, lastVersion string) (*AiRouteDataExport, error) {
	rst, err := e.versionControlManager.ExportConfig(ctx, ConfigTopicProductAIRoute, e.AIRouteGenerator)
	if err != nil {
		return nil, err
	}

	if rst.DataWithoutVersion == nil {
		return nil, fmt.Errorf("AIRouteGenerator.DataWithoutVersion is nil")
	}

	conf, ok := rst.DataWithoutVersion.(*AiRouteDataExport)
	if !ok {
		return nil, fmt.Errorf("convert AIRouteGenerator.DataWithoutVersion to AiRouteDataExport is error")
	}

	if conf.Version == lastVersion {
		return nil, nil
	}

	return conf, nil
}

// AIRouteGenerator generates AI route configuration data
func (e *AIRouteExporter) AIRouteGenerator(ctx context.Context) (*iversion_control.ExportData, error) {
	apiKeys, err := e.apiKeyStorager.FetchAPIKeyList(ctx, &icluster_conf.APIKeyFilter{})
	if err != nil {
		return nil, fmt.Errorf("fetch api keys error: %s", err.Error())
	}

	entities, err := e.entityStorager.FetchEntityList(ctx, &quota.EntityFilter{})
	if err != nil {
		return nil, fmt.Errorf("fetch entities error: %s", err.Error())
	}
	entityMap := make(map[string]*quota.EntityParam)
	for _, entity := range entities {
		if entity.EntityID != nil {
			entityMap[*entity.EntityID] = entity
		}
	}

	globalOwner := shared.RouteRulesTypeGlobal
	globalRouteRules, err := e.routeRulesStorager.FetchRouteRules(ctx, shared.RouteRulesTypeGlobal, &globalOwner)
	if err != nil {
		return nil, fmt.Errorf("fetch global route rules error: %s", err.Error())
	}

	routeRules := make(map[string]*RouteTableExport)
	bindings := make(map[string][]string)

	for _, apiKey := range apiKeys {
		if apiKey.Key == nil {
			continue
		}

		var bindingList []string

		// API-Key level route table
		if apiKey.RouteRulesID != nil {
			routeRulesParam, err := e.routeRulesStorager.FetchRouteRulesByID(ctx, *apiKey.RouteRulesID)
			if err != nil {
				return nil, err
			}
			if routeRulesParam != nil && routeRulesParam.Enabled != nil && *routeRulesParam.Enabled {
				tableKey := fmt.Sprintf("apikey_%s", *apiKey.Key)
				routeRules[tableKey] = convertToRouteTableExport(shared.RouteRulesTypeAPIKey, *apiKey.Key, routeRulesParam.Rules)
				bindingList = append(bindingList, tableKey)
			}
		}

		// Entity level route table
		if apiKey.EntityID != nil && *apiKey.EntityID != "" {
			entity, exists := entityMap[*apiKey.EntityID]
			if exists && entity.RouteRulesID != nil {
				routeRulesParam, err := e.routeRulesStorager.FetchRouteRulesByID(ctx, *entity.RouteRulesID)
				if err != nil {
					return nil, err
				}
				if routeRulesParam != nil && routeRulesParam.Enabled != nil && *routeRulesParam.Enabled && entity.Name != nil {
					tableKey := fmt.Sprintf("entity_%s", *entity.Name)
					routeRules[tableKey] = convertToRouteTableExport(shared.RouteRulesTypeEntity, *entity.Name, routeRulesParam.Rules)
					bindingList = append(bindingList, tableKey)
				}
			}
		}

		// Global level route table
		if globalRouteRules != nil && globalRouteRules.Enabled != nil && *globalRouteRules.Enabled {
			tableKey := "global_default"
			routeRules[tableKey] = convertToRouteTableExport(shared.RouteRulesTypeGlobal, shared.RouteRulesTypeGlobal, globalRouteRules.Rules)
			bindingList = append(bindingList, tableKey)
		}

		if len(bindingList) > 0 {
			bindings[*apiKey.Key] = bindingList
		}
	}

	conf := &AiRouteDataExport{
		RouteRules:               routeRules,
		ApikeyRouteTableBindings: bindings,
	}
	conf.UpdateVersion(iversion_control.ZeroVersion)

	return &iversion_control.ExportData{
		Topic:              ConfigTopicProductAIRoute,
		DataWithoutVersion: conf,
	}, nil
}

func convertToRouteTableExport(ruleType string, owner string, rules []*shared.AiRouteRuleParam) *RouteTableExport {
	table := &RouteTableExport{
		Type:  ruleType,
		Owner: owner,
		Rules: []*RouteRuleExport{},
	}

	for _, rule := range rules {
		exportRule := &RouteRuleExport{}
		if rule.Name != nil {
			exportRule.Name = *rule.Name
		}
		if rule.Cond != nil {
			exportRule.Cond = *rule.Cond
		}

		for _, target := range rule.Targets {
			exportTarget := &AiRouteTargetExport{}
			if target.ClusterName != nil {
				exportTarget.ClusterName = *target.ClusterName
			}
			if target.Model != nil {
				exportTarget.Model = *target.Model
			}
			if target.Weight != nil {
				exportTarget.Weight = *target.Weight
			}
			exportRule.Targets = append(exportRule.Targets, exportTarget)
		}

		for _, fallback := range rule.Fallbacks {
			exportFallback := &AiRouteFallbackExport{}
			if fallback.ClusterName != nil {
				exportFallback.ClusterName = *fallback.ClusterName
			}
			if fallback.Model != nil {
				exportFallback.Model = *fallback.Model
			}
			exportRule.Fallbacks = append(exportRule.Fallbacks, exportFallback)
		}

		table.Rules = append(table.Rules, exportRule)
	}

	return table
}
