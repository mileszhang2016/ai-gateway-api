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

package api_key

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"

	"github.com/yf-networks/ai-gateway-api/lib"
	"github.com/yf-networks/ai-gateway-api/lib/xerror"
	"github.com/yf-networks/ai-gateway-api/lib/xreq"
	"github.com/yf-networks/ai-gateway-api/model/iauth"
	"github.com/yf-networks/ai-gateway-api/model/ibasic"
	"github.com/yf-networks/ai-gateway-api/model/icluster_conf"
	"github.com/yf-networks/ai-gateway-api/model/quota"
	"github.com/yf-networks/ai-gateway-api/model/shared"
	"github.com/yf-networks/ai-gateway-api/stateful"
	"github.com/yf-networks/ai-gateway-api/stateful/container"
)

const defaultProductName = "AI_product"

func defaultProduct() *ibasic.Product {
	return &ibasic.Product{Name: defaultProductName}
}

var _ xreq.Handler = APIKeyCreateAction

var APIKeyCreateRoute = &xreq.Endpoint{
	Path:       "/api-keys",
	Method:     http.MethodPost,
	Handler:    xreq.Convert(APIKeyCreateAction),
	Authorizer: iauth.FA(iauth.FeatureAPIKey, iauth.ActionCreate),
}

func APIKeyCreateAction(req *http.Request) (interface{}, error) {
	param := &icluster_conf.APIKeyParam{}
	if err := xreq.BindJSON(req, param); err != nil {
		return nil, err
	}

	return APIKeyCreateProcess(req.Context(), param, defaultProduct())
}

func APIKeyCreateProcess(ctx context.Context, param *icluster_conf.APIKeyParam, product *ibasic.Product) (*icluster_conf.APIKeyParam, error) {
	if param.Key == nil || *param.Key == "" {
		generatedKey, err := generateAPIKeyValue(product.Name)
		if err != nil {
			return nil, err
		}
		param.Key = &generatedKey
	}

	if err := checkCreateAPIKey(param, product.Name); err != nil {
		return nil, xerror.WrapParamError(err)
	}

	if param.EntityID != nil && *param.EntityID != "" {
		entity, err := container.EntityStorager.FetchEntity(ctx, &quota.EntityFilter{EntityID: param.EntityID})
		if err != nil {
			return nil, err
		}
		if entity == nil {
			return nil, xerror.WrapRecordNotExist("Entity")
		}
	}

	if param.Enable == nil {
		enabled := true
		param.Enable = &enabled
	}

	apiKeyID, err := generateAPIKeyID(ctx, product.Name)
	if err != nil {
		return nil, err
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

	err = container.APIKeyManager.CreateAPIKey(ctx, &icluster_conf.APIKeyParam{
		ID:              lib.PString(apiKeyID),
		Enable:          param.Enable,
		Key:             param.Key,
		Description:     param.Description,
		UnlimitedQuota:  param.UnlimitedQuota,
		ExpiredTime:     param.ExpiredTime,
		Models:          param.Models,
		Subnet:          param.Subnet,
		EntityID:        param.EntityID,
		ProductName:     &product.Name,
		QuotaPlan:       param.QuotaPlan,
		RateLimitPolicy: param.RateLimitPolicy,
	})

	if err == nil && param.Key != nil && param.QuotaPlan != nil &&
		(param.QuotaPlan.Unlimited == nil || !*param.QuotaPlan.Unlimited) &&
		param.QuotaPlan.Quota != nil {
		redisKey := stateful.AIUsedQuotaKey(*param.Key)
		currentValue, errGet := stateful.DefaultClientSet.RedisClient.GetInt64(redisKey)
		if errGet != nil {
			_, err = stateful.DefaultClientSet.RedisClient.IncrBy(redisKey, *param.QuotaPlan.Quota)
		} else {
			delta := *param.QuotaPlan.Quota - currentValue
			_, err = stateful.DefaultClientSet.RedisClient.IncrBy(redisKey, delta)
		}
	}

	if err != nil {
		return nil, err
	}

	return container.APIKeyManager.FetchAPIKey(ctx, &icluster_conf.APIKeyFilter{
		ID:          &apiKeyID,
		ProductName: &product.Name,
	})
}

// generateAPIKeyID generates a unique API-Key ID with format "api-key-{sequence}"
func generateAPIKeyID(ctx context.Context, productName string) (string, error) {
	// Get all API keys to find the max sequence number
	list, err := container.APIKeyManager.FetchAPIKeyList(ctx, &icluster_conf.APIKeyFilter{
		ProductName: &productName,
	})
	if err != nil {
		return "", err
	}

	maxSeq := 0
	for _, apiKey := range list {
		if apiKey.ID != nil {
			var seq int
			if _, err := fmt.Sscanf(*apiKey.ID, "api-key-%d", &seq); err == nil {
				if seq > maxSeq {
					maxSeq = seq
				}
			}
		}
	}

	return fmt.Sprintf("api-key-%d", maxSeq+1), nil
}

// generateAPIKeyValue generates a random API-Key value with format "{productName}-{randomChars}"
// randomChars consists of uppercase letters, lowercase letters, and numbers
func generateAPIKeyValue(productName string) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	const keyLength = 24

	b := make([]byte, keyLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return fmt.Sprintf("%s-%s", productName, string(b)), nil
}
