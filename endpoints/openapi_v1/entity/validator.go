// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
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
	"github.com/rainway-ai-gateway/ai-gateway-api/lib/validate"
	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/entity"
)

// validateEntityParam performs centralized validation on entity request parameters.
// When requireNameType is true, Name and Type are required (used by create).
func validateEntityParam(param *entity.EntityParam, requireNameType bool) error {
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

// validateTypeImmutable 执行 Entity type 不可变约束（api-define entities.md §2.4/§2.5）：
// type 创建后固定，全量/部分更新携带与库中不同的值时拒绝。
// 携带与库中相同的值放行（GET→修改→PUT 回环必须可用）；paramType 为 nil（未携带）不触发。
func validateTypeImmutable(paramType, existingType *string) error {
	if paramType != nil && existingType != nil && *paramType != *existingType {
		return xerror.WrapParamErrorWithMsg("type is immutable, cannot be modified after creation")
	}
	return nil
}
