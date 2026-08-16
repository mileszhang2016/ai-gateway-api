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

package api_key

import (
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/validate"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/api_key"
)

// checkCreateAPIKey validates parameters for creating a new API key.
func checkCreateAPIKey(param *api_key.APIKeyParam, productName string) error {
	if param.Description == nil || *param.Description == "" {
		return xerror.WrapParamErrorWithMsg("description is required")
	}
	if err := validate.APIKeyDescription(*param.Description); err != nil {
		return err
	}

	if err := checkAllowSubnet(param.Subnet); err != nil {
		return err
	}

	if err := checkKey(param.Key, productName); err != nil {
		return err
	}

	if err := validate.ExpiredTime(param.ExpiredTime); err != nil {
		return err
	}

	if err := validate.RateLimitPolicy(param.RateLimitPolicy); err != nil {
		return err
	}

	if err := validate.QuotaPlan(param.QuotaPlan); err != nil {
		return err
	}

	if err := validate.RouteRules(param.RouteRules); err != nil {
		return err
	}

	return nil
}

// checkUpdateAPIKey validates parameters for updating an existing API key.
func checkUpdateAPIKey(param *api_key.APIKeyParam, productName string) error {
	if param.Description != nil {
		if err := validate.Description(*param.Description, validate.MaxAPIDescriptionLength, "description"); err != nil {
			return err
		}
	}

	if err := checkAllowSubnet(param.Subnet); err != nil {
		return err
	}

	if param.Key != nil {
		if err := validate.APIKeyValue(*param.Key); err != nil {
			return err
		}
	}

	if err := validate.ExpiredTime(param.ExpiredTime); err != nil {
		return err
	}

	if err := validate.RateLimitPolicy(param.RateLimitPolicy); err != nil {
		return err
	}

	if err := validate.QuotaPlan(param.QuotaPlan); err != nil {
		return err
	}

	if err := validate.RouteRules(param.RouteRules); err != nil {
		return err
	}

	return nil
}

// checkKey validates the API key value.
func checkKey(key *string, productName string) error {
	if key == nil || *key == "" {
		return xerror.WrapParamErrorWithMsg("Must set key")
	}
	return validate.APIKeyValue(*key)
}

// checkAllowSubnet validates CIDR subnet format.
func checkAllowSubnet(cidrs []string) error {
	for _, cidr := range cidrs {
		if err := validate.CIDR(cidr); err != nil {
			return err
		}
	}
	return nil
}

// checkFullUpdateAPIKey validates parameters for full updating an existing API key.
func checkFullUpdateAPIKey(param *api_key.APIKeyParam, productName string) error {
	if param.Description == nil || *param.Description == "" {
		return xerror.WrapParamErrorWithMsg("description is required")
	}
	if err := validate.APIKeyDescription(*param.Description); err != nil {
		return err
	}

	return checkUpdateAPIKey(param, productName)
}
