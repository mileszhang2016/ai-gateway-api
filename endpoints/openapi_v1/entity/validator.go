// Copyright(c) 2026 Beijing Yingfei Networks Technology Co.Ltd.
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

package entity

import (
	"github.com/yf-networks/ai-gateway-api/lib/validate"
	"github.com/yf-networks/ai-gateway-api/lib/xerror"
	"github.com/yf-networks/ai-gateway-api/model/quota"
)

// validateEntityParam performs centralized validation on entity request parameters.
// When requireNameType is true, Name and Type are required (used by create).
func validateEntityParam(param *quota.EntityParam, requireNameType bool) error {
	if requireNameType {
		if param.Name == nil || *param.Name == "" {
			return xerror.WrapParamErrorWithMsg("name is required")
		}
		if param.Type == nil || *param.Type == "" {
			return xerror.WrapParamErrorWithMsg("type is required")
		}
	}

	if param.Name != nil && *param.Name != "" {
		if err := validate.EntityName(*param.Name); err != nil {
			return err
		}
	}

	if param.Type != nil {
		if err := validate.EntityTypeName(*param.Type); err != nil {
			return err
		}
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
