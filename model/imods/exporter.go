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

package imods

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iai_route"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iversion_control"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/quota"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/shared"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful"
)

// ConfigTopicProductAPIKeyRule is the configuration topic for API key rules
const ConfigTopicProductAPIKeyRule = "mod_api_key_rule"

const (
	TokenStatusEnabled   = 1
	TokenStatusDisabled  = 2
	TokenStatusExpired   = 3
	TokenStatusExhausted = 4
)

// ModAPIKeyRuleConf defines the configuration structure for API key rules module
type ModAPIKeyRuleConf struct {
	Version    *string                          `json:"version"`
	Config     map[string][]*TokenRuleFile      `json:"config"`
	QuotaPlans map[string][]*QuotaPlan          `json:"QuotaPlans"`
	Tokens     map[string]map[string]*TokenFile `json:"tokens"`
}

type TokenRuleFile struct {
	Cond   *string
	Action *ActionFile
}

type ApikeyTag struct {
	TagName  string //eg entity.type
	TagValue string //eg entity.name
}

type ActionFile struct {
	Cmd string
}

type QuotaPlan struct {
	Id          string
	Unlimited   bool
	PassNoQuota bool
	RedisKey    string
	CreateTime  int64
	ExpiredTime int64 // -1 means never expired
	Quota       int64 // 配额总量
	ResetMode   int   // 0 – 非周期性；1 – 周期性的配额包
}

type TokenFile struct {
	Key            string  `json:"key"`
	Enabled        int     `json:"enabled"`
	Status         int     `json:"status"`
	Name           string  `json:"name"`
	UpdateTime     int64   `json:"update_time"`
	ExpiredTime    int64   `json:"expired_time"` // -1 means never expired
	UnlimitedQuota bool    `json:"unlimited_quota"`
	Models         *string `json:"allow_models"` // allowed models
	BlockModels    *string `json:"block_models"` // blocked models
	Subnet         *string `json:"subnet"`       // allowed subnet
	Tags           []ApikeyTag
	QuotaPlans     []string `json:"quota_plans"` // quotaPlan IDs
	models         []string
	blockModels    []string
	subnet         []*net.IPNet
}

// UpdateVersion updates the configuration version
func (conf *ModAPIKeyRuleConf) UpdateVersion(version string) error {
	conf.Version = &version
	return nil
}

// ConfigExport exports API key rule configuration for BFE
func (rcm *APIKeyRuleManager) ConfigExport(ctx context.Context, lastVersion string) (*ModAPIKeyRuleConf, error) {
	// Export configuration using version control manager
	rst, err := rcm.versionControlManager.ExportConfig(ctx, ConfigTopicProductAPIKeyRule, rcm.APIKeyRuleGenerator)
	if err != nil {
		return nil, err
	}

	if rst.DataWithoutVersion == nil {
		return nil, fmt.Errorf("APIKeyRuleGenerator.DataWithoutVersion is nil")
	}

	// Convert exported data to configuration structure
	conf, ok := rst.DataWithoutVersion.(*ModAPIKeyRuleConf)
	if ok {
		// Return nil if version hasn't changed
		if *conf.Version == lastVersion {
			return nil, nil
		}

		return conf, nil
	}

	return nil, fmt.Errorf("convert APIKeyRuleGenerator.DataWithoutVersion to ModAPIKeyRuleConf is error")
}

func (rlm *APIKeyRuleManager) FormatAIRouteAPIKeyRules(ctx context.Context, productName string) ([]*APIKeyRule, error) {
	aiRouteRules, err := rlm.aiRouteStorager.FetchAIRouteRules(ctx, &iai_route.AIRouteFilter{
		ProductName: &productName,
	})
	if err != nil {
		return nil, err
	}

	var ruleResult []*APIKeyRule
	for _, rule := range aiRouteRules {
		cond := iai_route.BuildAIRouteCond(ctx, rule.Basic)

		ruleResult = append(ruleResult, &APIKeyRule{
			Cond: cond,
			Actions: []Action{
				{
					Cmd: APIKeyActionCMD,
				},
			},
			ProductName: stateful.DefaultConfig.RunTime.AIRouteInnerProductName,
		})
	}

	ruleResult = append(ruleResult, &APIKeyRule{
		Cond: "default_t()",
		Actions: []Action{
			{
				Cmd: APIKeyActionCMD,
			},
		},
		ProductName: stateful.DefaultConfig.RunTime.AIRouteInnerProductName,
	})

	return ruleResult, nil
}

func (rlm *APIKeyRuleManager) buildAIRouteAPIKeyRules(ctx context.Context) (map[string]*APIKey, error) {
	aiProductName := stateful.DefaultConfig.RunTime.AIRouteInnerProductName
	rules, err := rlm.FormatAIRouteAPIKeyRules(ctx, aiProductName)
	if err != nil {
		return nil, fmt.Errorf("read ai route api key rule is error:%s", err.Error())
	}

	product2config := make(map[string]*APIKey)
	if len(rules) > 0 {
		product2config[aiProductName] = &APIKey{
			Rules: rules,
		}
	}

	return product2config, nil
}

// APIKeyRuleGenerator generates API key rules and token information for BFE configuration
func (rlm *APIKeyRuleManager) APIKeyRuleGenerator(ctx context.Context) (*iversion_control.ExportData, error) {
	collectedQuotaPlans := make(map[string][]*QuotaPlan)

	product2config, err := rlm.buildAIRouteAPIKeyRules(ctx)
	if err != nil {
		return nil, err
	}

	productName2Config := make(map[string][]*TokenRuleFile)
	for productName, productConfig := range product2config {
		if len(productConfig.Rules) > 0 {
			productName2Config[productName] = convertAPIKeyRulesToBfeRules(productConfig.Rules)
		}
	}

	apiKeyList, err := rlm.apiKeyStorager.FetchAPIKeyList(ctx, &icluster_conf.APIKeyFilter{})
	if err != nil {
		return nil, err
	}

	// Populate QuotaPlan for status calculation. The raw storager does not load
	// associated objects, but GetRemainingQuota needs QuotaPlan.Quota to decide
	// whether the token is exhausted.
	for _, one := range apiKeyList {
		if one.QuotaPlanID != nil && rlm.quotaPlanStorager != nil {
			quotaPlan, err := rlm.quotaPlanStorager.FetchQuotaPlan(ctx, &quota.QuotaPlanFilter{ID: one.QuotaPlanID})
			if err != nil {
				return nil, err
			}
			if quotaPlan != nil {
				one.QuotaPlan = &shared.QuotaPlanParam{
					Unlimited:             quotaPlan.Unlimited,
					PassWhenNoEnoughQuota: quotaPlan.PassWhenNoEnoughQuota,
					Quota:                 quotaPlan.Quota,
					Unit:                  quotaPlan.Unit,
					ResetPeriod:           quotaPlan.ResetPeriod,
				}
			}
		}
	}

	apiKey2Config := make(map[string]map[string]*TokenFile)
	for _, one := range apiKeyList {
		if _, ok := apiKey2Config[*one.ProductName]; !ok {
			items := make(map[string]*TokenFile)
			apiKey2Config[*one.ProductName] = items
		}

		expiredTime := int64(UnlimitedQuota)
		if one.ExpiredTime != nil {
			expiredTime = *one.ExpiredTime
		}

		status := TokenStatusDisabled
		var remainQuota float64
		if one.Enable != nil && *one.Enable {
			isUnlimited := one.UnlimitedQuota != nil && *one.UnlimitedQuota
			if !isUnlimited {
				status = TokenStatusEnabled
				remainingQuota, err := icluster_conf.GetRemainingQuota(one)
				if err != nil {
					return nil, err
				}

				if remainingQuota != nil {
					remainQuota = *remainingQuota
					if expiredTime != int64(UnlimitedQuota) && time.Now().Local().Unix() >= expiredTime {
						status = TokenStatusExpired
					} else if remainQuota <= 0 {
						status = TokenStatusExhausted
					}
				}
			} else {
				status = TokenStatusEnabled
			}
		}

		items := apiKey2Config[*one.ProductName]

		tokenFile := &TokenFile{
			Key:            *one.Key,
			Enabled:        2,
			Status:         status,
			Name:           *one.ID,
			ExpiredTime:    expiredTime,
			UpdateTime:     one.KeyCreateAt.Unix(),
			UnlimitedQuota: one.UnlimitedQuota != nil && *one.UnlimitedQuota,
		}
		if one.Enable != nil && *one.Enable {
			tokenFile.Enabled = 1
		}

		// Collect allow_models and block_models from entity hierarchy
		var entityAllowModels []string
		var entityBlockModels []string
		if one.EntityID != nil && *one.EntityID != "" && rlm.entityStorager != nil {
			entityAllowModels, entityBlockModels, err = rlm.fetchEntityModelHierarchy(ctx, *one.EntityID)
			if err != nil {
				return nil, fmt.Errorf("fetch entity model hierarchy error: %s", err.Error())
			}
		}

		// Collect api-key's non-empty, non-* allow_models
		var apiKeyAllowModels []string
		for _, m := range one.Models {
			if m != "" && m != "*" {
				apiKeyAllowModels = append(apiKeyAllowModels, m)
			}
		}

		// entityAllowModels is already the intersection of all non-* entity allow_models
		// Compute the final intersection: api-key models ∩ entity models
		var finalAllowModels []string
		if len(apiKeyAllowModels) == 0 && len(entityAllowModels) == 0 {
			// Rule 1: both are empty or * → allow all
			modelsStr := ""
			tokenFile.Models = &modelsStr
		} else {
			// Rule 2: compute intersection of all non-empty, non-* models
			if len(apiKeyAllowModels) == 0 {
				finalAllowModels = entityAllowModels
			} else if len(entityAllowModels) == 0 {
				finalAllowModels = apiKeyAllowModels
			} else {
				finalAllowModels = intersectSlices(apiKeyAllowModels, entityAllowModels)
			}

			if len(finalAllowModels) == 0 {
				// Rule 3: intersection empty → disable the token, models empty
				modelsStr := ""
				tokenFile.Models = &modelsStr
				tokenFile.Enabled = 2
			} else {
				modelsStr := strings.Join(finalAllowModels, ",")
				tokenFile.Models = &modelsStr
			}
		}

		// Merge block_models from entity hierarchy (union of all)
		if len(entityBlockModels) > 0 {
			blockModelsStr := strings.Join(entityBlockModels, ",")
			if blockModelsStr == "*" {
				blockModelsStr = ""
			}
			tokenFile.BlockModels = &blockModelsStr
		}

		if len(one.Subnet) > 0 {
			subnetStr := strings.Join(one.Subnet, ",")
			if subnetStr == "*" {
				subnetStr = ""
			}
			tokenFile.Subnet = &subnetStr
		}

		quotaPlanIDs, tags, err := rlm.fetchQuotaPlansWithEntityHierarchy(ctx, one, *one.ProductName, collectedQuotaPlans)
		if err != nil {
			return nil, fmt.Errorf("fetch quota plans error: %s", err.Error())
		}
		tokenFile.QuotaPlans = quotaPlanIDs
		tokenFile.Tags = tags

		// Defense: if all related quota plans are unlimited and QuotaPlans is empty,
		// treat this token as unlimited to avoid BFE load failure.
		if !tokenFile.UnlimitedQuota && len(tokenFile.QuotaPlans) == 0 {
			tokenFile.UnlimitedQuota = true
		}

		items[*one.Key] = tokenFile
		apiKey2Config[*one.ProductName] = items
	}

	conf := &ModAPIKeyRuleConf{
		Config:     productName2Config,
		QuotaPlans: collectedQuotaPlans,
		Tokens:     apiKey2Config,
	}

	conf.UpdateVersion(iversion_control.ZeroVersion)

	return &iversion_control.ExportData{
		Topic:              ConfigTopicProductAPIKeyRule,
		DataWithoutVersion: conf,
	}, nil
}

// convertAPIKeyRulesToBfeRules converts internal API key rules to BFE format
func convertAPIKeyRulesToBfeRules(oldRules []*APIKeyRule) []*TokenRuleFile {
	exportRules := make([]*TokenRuleFile, len(oldRules))
	for i, rule := range oldRules {
		newRule := &TokenRuleFile{
			Cond: &rule.Cond,
		}

		if len(rule.Actions) > 0 {
			newRule.Action = &ActionFile{
				Cmd: rule.Actions[0].Cmd,
			}
		}

		exportRules[i] = newRule
	}
	return exportRules
}

func (rlm *APIKeyRuleManager) fetchQuotaPlansWithEntityHierarchy(ctx context.Context, apiKey *icluster_conf.APIKeyParam, productName string, collectedQuotaPlans map[string][]*QuotaPlan) ([]string, []ApikeyTag, error) {
	quotaPlanIDs := make([]string, 0)
	tags := make([]ApikeyTag, 0)

	if apiKey.QuotaPlanID != nil && rlm.quotaPlanStorager != nil {
		quotaPlan, err := rlm.quotaPlanStorager.FetchQuotaPlan(ctx, &quota.QuotaPlanFilter{ID: apiKey.QuotaPlanID})
		if err != nil {
			return nil, nil, err
		}
		if quotaPlan != nil {
			// Skip API-Key own unlimited quota plans: they do not need to be
			// referenced in the token's quota_plans list.
			if quotaPlan.Unlimited == nil || !*quotaPlan.Unlimited {
				qp := convertQuotaPlanToExport(quotaPlan, *apiKey.Key, *apiKey.Key)
				if !containsQuotaPlan(collectedQuotaPlans, productName, qp.Id) {
					if _, ok := collectedQuotaPlans[productName]; !ok {
						collectedQuotaPlans[productName] = make([]*QuotaPlan, 0)
					}
					collectedQuotaPlans[productName] = append(collectedQuotaPlans[productName], qp)
				}
				quotaPlanIDs = append(quotaPlanIDs, qp.Id)
			}
		}
	}

	if apiKey.EntityID != nil && *apiKey.EntityID != "" && rlm.entityStorager != nil {
		entity, err := rlm.entityStorager.FetchEntity(ctx, &quota.EntityFilter{EntityID: apiKey.EntityID})
		if err != nil {
			return nil, nil, err
		}
		if entity != nil {
			entityQuotaPlanIDs, entityTags, err := rlm.fetchEntityQuotaPlanHierarchy(ctx, entity, productName, collectedQuotaPlans)
			if err != nil {
				return nil, nil, err
			}
			quotaPlanIDs = append(quotaPlanIDs, entityQuotaPlanIDs...)
			tags = append(tags, entityTags...)
		}
	}

	return quotaPlanIDs, tags, nil
}

func (rlm *APIKeyRuleManager) fetchEntityQuotaPlanHierarchy(ctx context.Context, entity *quota.EntityParam, productName string, collectedQuotaPlans map[string][]*QuotaPlan) ([]string, []ApikeyTag, error) {
	quotaPlanIDs := make([]string, 0)
	tags := make([]ApikeyTag, 0)

	if entity.EntityID != nil && entity.Type != nil && entity.Name != nil {
		tags = append(tags, ApikeyTag{
			TagName:  *entity.Type,
			TagValue: *entity.Name,
		})
	}

	if entity.QuotaPlanID != nil && rlm.quotaPlanStorager != nil {
		quotaPlan, err := rlm.quotaPlanStorager.FetchQuotaPlan(ctx, &quota.QuotaPlanFilter{ID: entity.QuotaPlanID})
		if err != nil {
			return nil, nil, err
		}
		if quotaPlan != nil && entity.EntityID != nil {
			// Skip unlimited entity quota plans: they do not enforce any quota
			// and should not be referenced by the token.
			if quotaPlan.Unlimited == nil || !*quotaPlan.Unlimited {
				qp := convertQuotaPlanToExport(quotaPlan, *entity.EntityID, *entity.EntityID)
				if !containsQuotaPlan(collectedQuotaPlans, productName, qp.Id) {
					if _, ok := collectedQuotaPlans[productName]; !ok {
						collectedQuotaPlans[productName] = make([]*QuotaPlan, 0)
					}
					collectedQuotaPlans[productName] = append(collectedQuotaPlans[productName], qp)
				}
				quotaPlanIDs = append(quotaPlanIDs, qp.Id)
			}
		}
	}

	if entity.ParentID != nil && *entity.ParentID != "" && rlm.entityStorager != nil {
		parentEntity, err := rlm.entityStorager.FetchEntity(ctx, &quota.EntityFilter{EntityID: entity.ParentID})
		if err != nil {
			return nil, nil, err
		}
		if parentEntity != nil {
			parentQuotaPlanIDs, parentTags, err := rlm.fetchEntityQuotaPlanHierarchy(ctx, parentEntity, productName, collectedQuotaPlans)
			if err != nil {
				return nil, nil, err
			}
			quotaPlanIDs = append(quotaPlanIDs, parentQuotaPlanIDs...)
			tags = append(tags, parentTags...)
		}
	}

	return quotaPlanIDs, tags, nil
}

// fetchEntityModelHierarchy recursively collects AllowModels and BlockModels from the entity hierarchy
// Returns:
//   - intersectedAllowModels: intersection of all AllowModels from entities in the hierarchy (nil if any entity has empty AllowModels)
//   - unionBlockModels: union of all BlockModels from entities in the hierarchy
func (rlm *APIKeyRuleManager) fetchEntityModelHierarchy(ctx context.Context, entityID string) ([]string, []string, error) {
	entity, err := rlm.entityStorager.FetchEntity(ctx, &quota.EntityFilter{EntityID: &entityID})
	if err != nil {
		return nil, nil, err
	}
	if entity == nil {
		return nil, nil, nil
	}

	var allAllowModels [][]string
	var allBlockModels []string

	rlm.collectEntityModels(ctx, entity, &allAllowModels, &allBlockModels)

	return intersectAllowModels(allAllowModels), allBlockModels, nil
}

// collectEntityModels recursively walks the entity hierarchy to collect AllowModels and BlockModels
// If an entity's AllowModels or BlockModels contains "*", it is skipped (meaning allow/block all)
func (rlm *APIKeyRuleManager) collectEntityModels(ctx context.Context, entity *quota.EntityParam, allAllowModels *[][]string, allBlockModels *[]string) {
	if len(entity.AllowModels) > 0 && !containsStar(entity.AllowModels) {
		*allAllowModels = append(*allAllowModels, entity.AllowModels)
	}
	if len(entity.BlockModels) > 0 && !containsStar(entity.BlockModels) {
		*allBlockModels = append(*allBlockModels, entity.BlockModels...)
	}

	if entity.ParentID != nil && *entity.ParentID != "" && rlm.entityStorager != nil {
		parent, err := rlm.entityStorager.FetchEntity(ctx, &quota.EntityFilter{EntityID: entity.ParentID})
		if err != nil || parent == nil {
			return
		}
		rlm.collectEntityModels(ctx, parent, allAllowModels, allBlockModels)
	}
}

// containsStar checks if the string slice contains "*"
func containsStar(slice []string) bool {
	for _, s := range slice {
		if s == "*" {
			return true
		}
	}
	return false
}

// intersectAllowModels computes the intersection of multiple AllowModels slices
// Returns nil if there are no AllowModels configured at any entity level
func intersectAllowModels(allAllowModels [][]string) []string {
	if len(allAllowModels) == 0 {
		return nil
	}

	result := make([]string, len(allAllowModels[0]))
	copy(result, allAllowModels[0])

	for _, models := range allAllowModels[1:] {
		result = intersectSlices(result, models)
	}

	return result
}

// intersectSlices returns the intersection of two string slices
func intersectSlices(a, b []string) []string {
	set := make(map[string]bool)
	for _, s := range b {
		set[s] = true
	}

	var result []string
	for _, s := range a {
		if set[s] {
			result = append(result, s)
		}
	}
	return result
}

func containsQuotaPlan(collectedQuotaPlans map[string][]*QuotaPlan, productName, id string) bool {
	qpList, ok := collectedQuotaPlans[productName]
	if !ok {
		return false
	}
	for _, qp := range qpList {
		if qp.Id == id {
			return true
		}
	}
	return false
}

func convertQuotaPlanToExport(qp *quota.QuotaPlanParam, id string, redisKeyID string) *QuotaPlan {
	result := &QuotaPlan{
		Id:          id,
		RedisKey:    fmt.Sprintf("QUOTA_%s", redisKeyID),
		Unlimited:   qp.Unlimited != nil && *qp.Unlimited,
		PassNoQuota: qp.PassWhenNoEnoughQuota != nil && *qp.PassWhenNoEnoughQuota,
		ExpiredTime: -1,
	}
	if qp.Quota != nil {
		result.Quota = lib.QuotaToRedisValue(qp.Quota, qp.Unit)
	}
	if qp.ResetPeriod != nil {
		if *qp.ResetPeriod == "weekly" || *qp.ResetPeriod == "monthly" {
			result.ResetMode = 1
		} else {
			result.ResetMode = 0
		}
	} else {
		result.ResetMode = 0
	}
	if qp.CreateTime != nil {
		result.CreateTime = *qp.CreateTime
	}

	return result
}
