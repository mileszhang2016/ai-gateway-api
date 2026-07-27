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
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/yf-networks/ai-gateway-api/lib/xerror"
	"github.com/yf-networks/ai-gateway-api/model/icluster_conf"
	"github.com/yf-networks/ai-gateway-api/model/shared"
)

// checkCreateAPIKey validates parameters for creating a new API key
func checkCreateAPIKey(param *icluster_conf.APIKeyParam, productName string) error {
	if param.Description == nil || *param.Description == "" {
		return xerror.WrapParamErrorWithMsg("description is required")
	}

	if len(*param.Description) >= 512 {
		return xerror.WrapParamErrorWithMsg("description must be less than 512 characters")
	}

	if err := checkAllowSubnet(param.Subnet); err != nil {
		return err
	}

	if err := checkKey(param.Key, productName); err != nil {
		return err
	}

	if err := checkExpiredTime(param.ExpiredTime); err != nil {
		return err
	}

	if err := checkRateLimitPolicy(param.RateLimitPolicy); err != nil {
		return err
	}

	if err := checkQuotaPlan(param.QuotaPlan); err != nil {
		return err
	}

	return nil
}

// checkExpiredTime validates the expiration time format
func checkExpiredTime(expiredTime *int64) error {
	if expiredTime == nil {
		return nil
	}

	// -1 表示永不过期，其他值必须是有效的 Unix 时间戳
	if *expiredTime < -1 {
		return xerror.WrapParamErrorWithMsg(fmt.Sprintf("Invalid expired_time: %d", *expiredTime))
	}

	// 非-1的时间戳必须大于等于当前时间
	if *expiredTime != -1 {
		currentTime := time.Now().Unix()
		if *expiredTime < currentTime {
			return xerror.WrapParamErrorWithMsg(fmt.Sprintf("expired_time must be >= current time, got %d, current %d", *expiredTime, currentTime))
		}
	}

	return nil
}

// checkUpdateAPIKey validates parameters for updating an existing API key
func checkUpdateAPIKey(param *icluster_conf.APIKeyParam, productName string) error {
	if param.Description != nil && len(*param.Description) >= 512 {
		return xerror.WrapParamErrorWithMsg("description must be less than 512 characters")
	}

	if err := checkAllowSubnet(param.Subnet); err != nil {
		return err
	}

	if param.Key != nil {
		if !ValidateString(*param.Key) {
			return xerror.WrapParamErrorWithMsg(fmt.Sprintf("Allowed Characters: Uppercase/Lowercase Letters, Numbers, Hyphen (-) and Underscore (_)"))
		}
	}

	if err := checkExpiredTime(param.ExpiredTime); err != nil {
		return err
	}

	if err := checkRateLimitPolicy(param.RateLimitPolicy); err != nil {
		return err
	}

	if err := checkQuotaPlan(param.QuotaPlan); err != nil {
		return err
	}

	return nil
}

// checkKey validates the API key value
func checkKey(key *string, productName string) error {
	if key == nil || *key == "" {
		return xerror.WrapParamErrorWithMsg(fmt.Sprintf("Must set key"))
	}

	if !ValidateString(*key) {
		return xerror.WrapParamErrorWithMsg(fmt.Sprintf("Allowed Characters: Uppercase/Lowercase Letters, Numbers, Hyphen (-) and Underscore (_)"))
	}

	return nil
}

// ValidateString checks if a string contains only allowed characters
func ValidateString(s string) bool {
	for _, r := range s {
		if !isValidChar(r) {
			return false
		}
	}
	return true
}

// isValidChar checks if a rune is a valid character for API key
func isValidChar(c rune) bool {
	return (c >= 65 && c <= 90) || // A-Z
		(c >= 97 && c <= 122) || // a-z
		(c >= 48 && c <= 57) || // 0-9
		(c == 45) || // Hyphen (-)
		(c == '_') // Underscore (_)
}

// checkAllowSubnet validates CIDR subnet format
func checkAllowSubnet(cidrs []string) error {
	for _, cidr := range cidrs {
		if cidr == "*" {
			return nil
		}

		arr := strings.Split(cidr, "/")
		if len(arr) != 2 {
			return xerror.WrapParamErrorWithMsg(fmt.Sprintf("invalid subnet format:%s", cidr))
		}

		ip := net.ParseIP(arr[0])
		if ip == nil {
			return xerror.WrapParamErrorWithMsg(fmt.Sprintf("invalid subnet:%s", cidr))
		}

		_, _, err := net.ParseCIDR(ip.String() + "/" + arr[1])
		if err != nil {
			return xerror.WrapParamErrorWithMsg(fmt.Sprintf("invalid subnet:%s parse error:%s", cidr, err.Error()))
		}
	}

	return nil
}

func checkRateLimitPolicy(policy *shared.RateLimitPolicyParam) error {
	if policy == nil {
		return nil
	}

	if policy.Enabled != nil && *policy.Enabled {
		if policy.Rules == nil {
			return xerror.WrapParamErrorWithMsg("when rate_limit_policy.enabled is true, rules must be set")
		}
		hasTpm := len(policy.Rules.TpmConfigs) > 0
		hasRpm := len(policy.Rules.RpmConfigs) > 0
		hasConcurrency := policy.Rules.MaxConcurrency != nil && *policy.Rules.MaxConcurrency >= 0
		if !hasTpm && !hasRpm && !hasConcurrency {
			return xerror.WrapParamErrorWithMsg("when rate_limit_policy.enabled is true, at least one of rules.tpm, rules.rpm, or rules.max_concurrency(>=0) must be set")
		}
	}

	if policy.Rules != nil {
		for _, tpm := range policy.Rules.TpmConfigs {
			if tpm.WindowMinutes < 1 || tpm.WindowMinutes > 360 {
				return xerror.WrapParamErrorWithMsg(fmt.Sprintf("tpm window_minutes must be between 1 and 360, got %d", tpm.WindowMinutes))
			}

			if tpm.StepMinutes < 1 || tpm.StepMinutes > 360 {
				return xerror.WrapParamErrorWithMsg(fmt.Sprintf("tpm step_minutes must be between 1 and 360, got %d", tpm.StepMinutes))
			}

			if tpm.StepMinutes > tpm.WindowMinutes {
				return xerror.WrapParamErrorWithMsg(fmt.Sprintf("tpm step_minutes (%d) must be <= window_minutes (%d)", tpm.StepMinutes, tpm.WindowMinutes))
			}
		}

		for _, rpm := range policy.Rules.RpmConfigs {
			if rpm.WindowMinutes < 1 || rpm.WindowMinutes > 360 {
				return xerror.WrapParamErrorWithMsg(fmt.Sprintf("rpm window_minutes must be between 1 and 360, got %d", rpm.WindowMinutes))
			}
		}
	}

	return nil
}

func checkQuotaPlan(quotaPlan *shared.QuotaPlanParam) error {
	if quotaPlan == nil {
		return nil
	}

	if quotaPlan.Quota != nil && *quotaPlan.Quota < 0 {
		return xerror.WrapParamErrorWithMsg(fmt.Sprintf("quota must be >= 0, got %d", *quotaPlan.Quota))
	}

	return nil
}

// checkFullUpdateAPIKey validates parameters for full updating an existing API key
func checkFullUpdateAPIKey(param *icluster_conf.APIKeyParam, productName string) error {
	if param.Description == nil || *param.Description == "" {
		return xerror.WrapParamErrorWithMsg("description is required")
	}

	if len(*param.Description) >= 512 {
		return xerror.WrapParamErrorWithMsg("description must be less than 512 characters")
	}

	return checkUpdateAPIKey(param, productName)
}
