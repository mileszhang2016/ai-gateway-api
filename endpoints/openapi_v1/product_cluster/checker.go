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

package product_cluster

import (
	"fmt"
	"net"
	"regexp"
	"strconv"

	"github.com/yf-networks/ai-gateway-api/lib/xerror"
	"github.com/yf-networks/ai-gateway-api/model/icluster_conf"
)

// 定义错误类型
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

const (
	maxServiceNameLen = 255
	maxGroupLen       = 255
)

func checkLLMConfig(llmConfig *icluster_conf.LLMConfig) error {
	if llmConfig == nil {
		return nil
	}

	if llmConfig.ServiceName != nil && len(*llmConfig.ServiceName) > maxServiceNameLen {
		return xerror.WrapParamErrorWithMsg(fmt.Sprintf("llm_config.service_name length must be lower than %s", strconv.Itoa(maxServiceNameLen)))
	}

	if llmConfig.Group != nil && len(*llmConfig.Group) > maxGroupLen {
		return xerror.WrapParamErrorWithMsg(fmt.Sprintf("llm_config.group length must be lower than %s", strconv.Itoa(maxGroupLen)))
	}

	if llmConfig.ModelEndpoint != nil {
		switch llmConfig.ModelEndpoint.Schema {
		case "http":
		case "https":
		default:
			return xerror.WrapParamErrorWithMsg(fmt.Sprintf("llm_config.model_endpoint.schema must be http or https"))
		}
	}

	if llmConfig.Enable == nil {
		return xerror.WrapParamErrorWithMsg(fmt.Sprintf("Must set llm_config.enable"))
	}

	if llmConfig.Enable != nil && *llmConfig.Enable {
		if llmConfig.ServiceName == nil || *llmConfig.ServiceName == "" {
			return xerror.WrapParamErrorWithMsg(fmt.Sprintf("Must set llm_config.service_name"))
		}

		if llmConfig.ModelEndpoint == nil {
			return xerror.WrapParamErrorWithMsg(fmt.Sprintf("Must set llm_config.model_endpoint"))
		}

		if llmConfig.ModelEndpoint.URI == "" {
			return xerror.WrapParamErrorWithMsg(fmt.Sprintf("Must set llm_config.model_endpoint.uri"))
		}

		if len(llmConfig.Models) == 0 {
			return xerror.WrapParamErrorWithMsg(fmt.Sprintf("Must set llm_config.models"))
		}
	}

	return nil
}

// 校验函数
func CheckEPPServer(server *icluster_conf.EPPServer) error {
	if server == nil {
		return nil
	}

	// Compatible with api-server behavior:
	// - domain + port mode
	// - endpoints mode
	if (server.Domain == nil || *server.Domain == "") && len(server.Endpoints) == 0 {
		return nil
	}

	if server.Domain != nil && *server.Domain != "" {
		if server.Port == nil {
			return &ValidationError{Field: "port", Message: "Must be set port"}
		}
		if !isValidDomain(*server.Domain) {
			return &ValidationError{Field: "domain", Message: "invalid domain"}
		}
		if *server.Port < 1 || *server.Port > 65535 {
			return &ValidationError{Field: "port", Message: "port must be in 1-65535"}
		}
		return nil
	}

	if len(server.Endpoints) == 0 {
		return &ValidationError{Field: "endpoints", Message: "Must set at least one endpoint"}
	}

	for i, endpoint := range server.Endpoints {
		if endpoint == nil {
			return &ValidationError{Field: fmt.Sprintf("endpoints[%d]", i), Message: "endpoints must not be empty"}
		}

		// 校验 IP
		if endpoint.IP == nil {
			return &ValidationError{Field: fmt.Sprintf("endpoints[%d].ip", i), Message: "Must set ip"}
		}
		if *endpoint.IP == "" {
			return &ValidationError{Field: fmt.Sprintf("endpoints[%d].ip", i), Message: "Must set ip"}
		}

		if !isValidIP(*endpoint.IP) {
			return &ValidationError{Field: fmt.Sprintf("endpoints[%d].ip", i), Message: "invalid ip"}
		}

		// 校验 Port
		if endpoint.Port == nil {
			return &ValidationError{Field: fmt.Sprintf("endpoints[%d].port", i), Message: "Must set port in endpoint"}
		}
		if *endpoint.Port < 1 || *endpoint.Port > 65535 {
			return &ValidationError{Field: fmt.Sprintf("endpoints[%d].port", i), Message: "port must be in 1-65535 in endpoint"}
		}
	}

	return nil
}

// 校验 IP 地址格式
func isValidIP(ip string) bool {
	// 支持 IPv4 和 IPv6
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil
}

func isValidDomain(domain string) bool {
	domainRegex := regexp.MustCompile(`^(?i:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$`)
	return domainRegex.MatchString(domain)
}
