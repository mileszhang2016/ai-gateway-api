package models

import (
	"net/http"
	"strings"

	"github.com/yf-networks/ai-gateway-api/lib/xerror"
	"github.com/yf-networks/ai-gateway-api/lib/xreq"
	"github.com/yf-networks/ai-gateway-api/model/iauth"
	"github.com/yf-networks/ai-gateway-api/model/icluster_conf"
	"github.com/yf-networks/ai-gateway-api/stateful/container"
)

var _ xreq.Handler = ModelsListAction

var ModelsListRoute = &xreq.Endpoint{
	Path:       "/models",
	Method:     http.MethodGet,
	Handler:    xreq.Convert(ModelsListAction),
	Authorizer: iauth.FA(iauth.FeatureProduct, iauth.ActionRead),
}

type ModelsListReq struct {
	Service string `form:"service"`
}

type ModelMapping struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ServiceModel struct {
	Name          string         `json:"name"`
	DisplayName   string         `json:"display_name"`
	Models        []string       `json:"models"`
	ModelMappings []*ModelMapping `json:"model_mappings"`
}

type ModelsListResp struct {
	Services []*ServiceModel `json:"services"`
}

func ModelsListAction(req *http.Request) (interface{}, error) {
	var listReq ModelsListReq
	if err := xreq.BindForm(req, &listReq); err != nil {
		return nil, xerror.WrapParamError(err)
	}

	list, err := container.ClusterManager.FetchClusterList(req.Context(), &icluster_conf.ClusterFilter{})
	if err != nil {
		return nil, err
	}

	serviceMap := make(map[string]*ServiceModel)

	for _, cluster := range list {
		if cluster.LLMConfig == nil || cluster.LLMConfig.Enable == nil || !*cluster.LLMConfig.Enable {
			continue
		}

		if cluster.LLMConfig.ServiceName == nil || *cluster.LLMConfig.ServiceName == "" {
			continue
		}

		serviceName := *cluster.LLMConfig.ServiceName

		if listReq.Service != "" && serviceName != listReq.Service {
			continue
		}

		if _, ok := serviceMap[serviceName]; !ok {
			serviceMap[serviceName] = &ServiceModel{
				Name:          serviceName,
				DisplayName:   capitalize(serviceName),
				Models:        make([]string, 0),
				ModelMappings: make([]*ModelMapping, 0),
			}
		}

		sm := serviceMap[serviceName]

		for _, model := range cluster.LLMConfig.Models {
			found := false
			for _, m := range sm.Models {
				if m == model {
					found = true
					break
				}
			}
			if !found {
				sm.Models = append(sm.Models, model)
			}
		}

		for _, mapping := range cluster.LLMConfig.ModelMappings {
			if mapping.Key == nil || mapping.Value == nil {
				continue
			}
			found := false
			for _, mm := range sm.ModelMappings {
				if mm.Key == *mapping.Key {
					found = true
					break
				}
			}
			if !found {
				sm.ModelMappings = append(sm.ModelMappings, &ModelMapping{
					Key:   *mapping.Key,
					Value: *mapping.Value,
				})
			}
		}
	}

	services := make([]*ServiceModel, 0, len(serviceMap))
	for _, sm := range serviceMap {
		services = append(services, sm)
	}

	return &ModelsListResp{Services: services}, nil
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(string(part[0])) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, " ")
}