package api_key

import (
	"net/http"

	"github.com/yf-networks/ai-gateway-api/lib/xerror"
	"github.com/yf-networks/ai-gateway-api/lib/xreq"
	"github.com/yf-networks/ai-gateway-api/model/iauth"
	"github.com/yf-networks/ai-gateway-api/model/icluster_conf"
	"github.com/yf-networks/ai-gateway-api/model/quota"
	"github.com/yf-networks/ai-gateway-api/stateful/container"
)

// ResetQuotaRoute 重置配额路由
var ResetQuotaRoute = &xreq.Endpoint{
	Path:       "/api-keys/{id}/quota-plan/reset",
	Method:     http.MethodPost,
	Handler:    xreq.Convert(ResetQuotaAction),
	Authorizer: iauth.FAP(iauth.FeatureAPIKey, iauth.ActionUpdate),
}

type ResetQuotaReq struct {
	Quota *int64 `json:"quota,omitempty"`
}

var _ xreq.Handler = ResetQuotaAction

// ResetQuotaAction 重置配额操作
func ResetQuotaAction(req *http.Request) (interface{}, error) {
	oneReq, err := newReq4One(req)
	if err != nil {
		return nil, err
	}

	resetReq := &ResetQuotaReq{}
	if err := xreq.BindJSON(req, resetReq); err != nil {
		return nil, err
	}

	productName := defaultProductName

	// 获取 API Key
	apiKey, err := container.APIKeyManager.FetchAPIKey(req.Context(), &icluster_conf.APIKeyFilter{
		ID:          oneReq.ID,
		ProductName: &productName,
	})
	if err != nil {
		return nil, err
	}
	if apiKey == nil {
		return nil, xerror.WrapRecordNotExist("API-Key")
	}

	// 检查是否有关联的配额计划
	if apiKey.QuotaPlanID == nil {
		return nil, xerror.WrapParamErrorWithMsg("API-Key has no quota plan")
	}

	// 重置配额余额（不更新 last_reset_at，避免影响定期重置调度）
	err = container.QuotaPlanManager.ResetBalance(req.Context(), *apiKey.QuotaPlanID, resetReq.Quota, false)
	if err != nil {
		return nil, err
	}

	// 获取更新后的配额计划信息
	quotaPlan, err := container.QuotaPlanManager.FetchQuotaPlan(req.Context(), &quota.QuotaPlanFilter{
		ID: apiKey.QuotaPlanID,
	})
	if err != nil {
		return nil, err
	}

	// 获取余额信息
	balance, err := container.QuotaPlanManager.FetchQuotaBalance(req.Context(), *apiKey.QuotaPlanID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"api_key_id": *oneReq.ID,
		"quota_plan": quotaPlan,
		"balance":    balance,
		"message":    "quota reset successfully",
	}, nil
}
