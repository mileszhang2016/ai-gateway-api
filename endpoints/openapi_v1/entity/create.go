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

package entity

import (
	"context"
	"fmt"
	"net/http"

	"github.com/yf-networks/ai-gateway-api/lib"
	"github.com/yf-networks/ai-gateway-api/lib/xerror"
	"github.com/yf-networks/ai-gateway-api/lib/xreq"
	"github.com/yf-networks/ai-gateway-api/model/iauth"
	"github.com/yf-networks/ai-gateway-api/model/quota"
	"github.com/yf-networks/ai-gateway-api/model/shared"
	"github.com/yf-networks/ai-gateway-api/stateful"
	"github.com/yf-networks/ai-gateway-api/stateful/container"
)

var EntityCreateRoute = &xreq.Endpoint{
	Path:       "/entities",
	Method:     http.MethodPost,
	Handler:    xreq.Convert(EntityCreateAction),
	Authorizer: iauth.FA(iauth.FeatureEntity, iauth.ActionCreate),
}

func EntityCreateAction(req *http.Request) (interface{}, error) {
	param := &quota.EntityParam{}
	if err := xreq.BindJSON(req, param); err != nil {
		return nil, err
	}

	if param.Name == nil || *param.Name == "" {
		return nil, xerror.WrapParamErrorWithMsg("name is required")
	}

	if param.Type == nil || *param.Type == "" {
		return nil, xerror.WrapParamErrorWithMsg("type is required")
	}

	entityType, err := container.EntityTypeStorager.FetchEntityType(req.Context(), &quota.EntityTypeFilter{TypeName: param.Type})
	if err != nil {
		return nil, err
	}
	if entityType == nil {
		return nil, xerror.WrapParamErrorWithMsg("type does not exist")
	}

	if param.ParentID != nil && *param.ParentID != "" {
		parentEntity, err := container.EntityManager.FetchEntity(req.Context(), &quota.EntityFilter{EntityID: param.ParentID})
		if err != nil {
			return nil, err
		}
		if parentEntity == nil {
			return nil, xerror.WrapParamErrorWithMsg("parent_id does not exist")
		}
		if parentEntity.Type == nil {
			return nil, xerror.WrapParamErrorWithMsg("parent entity has no type")
		}
		parentType, err := container.EntityTypeStorager.FetchEntityType(req.Context(), &quota.EntityTypeFilter{TypeName: parentEntity.Type})
		if err != nil {
			return nil, err
		}
		if parentType == nil || parentType.Level == nil {
			return nil, xerror.WrapParamErrorWithMsg("parent entity type is invalid")
		}
		if *parentType.Level >= *entityType.Level {
			return nil, xerror.WrapParamErrorWithMsg("parent level must be lower than current level")
		}
	}

	existingEntities, err := container.EntityManager.FetchEntityList(req.Context(), &quota.EntityFilter{Type: param.Type, Name: param.Name})
	if err != nil {
		return nil, err
	}
	if len(existingEntities) > 0 {
		return nil, xerror.WrapDuplicateData("entity")
	}

	if param.RateLimitPolicy != nil && param.RateLimitPolicy.Enabled != nil && *param.RateLimitPolicy.Enabled {
		if param.RateLimitPolicy.Rules == nil {
			return nil, xerror.WrapParamErrorWithMsg("when rate_limit_policy.enabled is true, rules must be set")
		}
		hasTpm := len(param.RateLimitPolicy.Rules.TpmConfigs) > 0
		hasRpm := len(param.RateLimitPolicy.Rules.RpmConfigs) > 0
		hasConcurrency := param.RateLimitPolicy.Rules.MaxConcurrency != nil && *param.RateLimitPolicy.Rules.MaxConcurrency >= 0
		if !hasTpm && !hasRpm && !hasConcurrency {
			return nil, xerror.WrapParamErrorWithMsg("when rate_limit_policy.enabled is true, at least one of rules.tpm, rules.rpm, or rules.max_concurrency(>=0) must be set")
		}
	}

	if param.RateLimitPolicy != nil && param.RateLimitPolicy.Rules != nil {
		for _, tpm := range param.RateLimitPolicy.Rules.TpmConfigs {
			if tpm.WindowMinutes < 1 || tpm.WindowMinutes > 360 {
				return nil, xerror.WrapParamErrorWithMsg(fmt.Sprintf("tpm window_minutes must be between 1 and 360, got %d", tpm.WindowMinutes))
			}

			if tpm.StepMinutes < 1 || tpm.StepMinutes > 360 {
				return nil, xerror.WrapParamErrorWithMsg(fmt.Sprintf("tpm step_minutes must be between 1 and 360, got %d", tpm.StepMinutes))
			}

			if tpm.StepMinutes > tpm.WindowMinutes {
				return nil, xerror.WrapParamErrorWithMsg(fmt.Sprintf("tpm step_minutes (%d) must be <= window_minutes (%d)", tpm.StepMinutes, tpm.WindowMinutes))
			}
		}

		for _, rpm := range param.RateLimitPolicy.Rules.RpmConfigs {
			if rpm.WindowMinutes < 1 || rpm.WindowMinutes > 360 {
				return nil, xerror.WrapParamErrorWithMsg(fmt.Sprintf("rpm window_minutes must be between 1 and 360, got %d", rpm.WindowMinutes))
			}
		}
	}

	if param.EntityID == nil || *param.EntityID == "" {
		generatedID, err := generateEntityID(req.Context())
		if err != nil {
			return nil, err
		}
		param.EntityID = &generatedID
	}

	if param.QuotaPlan == nil {
		param.QuotaPlan = &shared.QuotaPlanParam{
			Unlimited:             lib.PBool(true),
			PassWhenNoEnoughQuota: lib.PBool(false),
		}
	}

	if param.RateLimitPolicy == nil {
		param.RateLimitPolicy = &shared.RateLimitPolicyParam{
			Enabled: lib.PBool(false),
			Rules: &shared.RateLimitRules{
				TpmConfigs:     []shared.TPMConfig{},
				RpmConfigs:     []shared.RPMConfig{},
				MaxConcurrency: lib.PInt(0),
			},
		}
	}

	if _, err := container.EntityManager.CreateEntity(req.Context(), param); err != nil {
		return nil, err
	}

	if param.EntityID != nil && param.QuotaPlan != nil &&
		(param.QuotaPlan.Unlimited == nil || !*param.QuotaPlan.Unlimited) &&
		param.QuotaPlan.Quota != nil {
		redisKey := stateful.AIUsedQuotaKey(*param.EntityID)
		currentValue, errGet := stateful.DefaultClientSet.RedisClient.GetInt64(redisKey)
		if errGet != nil {
			_, _ = stateful.DefaultClientSet.RedisClient.IncrBy(redisKey, *param.QuotaPlan.Quota)
		} else {
			delta := *param.QuotaPlan.Quota - currentValue
			_, _ = stateful.DefaultClientSet.RedisClient.IncrBy(redisKey, delta)
		}
	}

	if param.EntityID != nil && param.QuotaPlan != nil &&
		param.QuotaPlan.Unlimited != nil && *param.QuotaPlan.Unlimited {
		redisKey := stateful.AIUsedQuotaKey(*param.EntityID)
		defaultQuota := int64(100000000)
		currentValue, errGet := stateful.DefaultClientSet.RedisClient.GetInt64(redisKey)
		if errGet != nil {
			_, _ = stateful.DefaultClientSet.RedisClient.IncrBy(redisKey, defaultQuota)
		} else {
			delta := defaultQuota - currentValue
			_, _ = stateful.DefaultClientSet.RedisClient.IncrBy(redisKey, delta)
		}
	}

	return container.EntityManager.FetchEntity(req.Context(), &quota.EntityFilter{EntityID: param.EntityID})
}

func generateEntityID(ctx context.Context) (string, error) {
	list, err := container.EntityManager.FetchEntityList(ctx, &quota.EntityFilter{})
	if err != nil {
		return "", err
	}

	maxSeq := 0
	for _, entity := range list {
		if entity.EntityID != nil {
			var seq int
			if _, err := fmt.Sscanf(*entity.EntityID, "entity-%d", &seq); err == nil {
				if seq > maxSeq {
					maxSeq = seq
				}
			}
		}
	}

	return fmt.Sprintf("entity-%d", maxSeq+1), nil
}
