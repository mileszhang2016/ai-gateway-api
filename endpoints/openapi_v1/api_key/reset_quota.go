package api_key

import (
	"net/http"
	"strings"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/quota"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

// ResetQuotaRoute 重置配额路由
var ResetQuotaRoute = &xreq.Endpoint{
	Path:       "/api-keys/{id}/quota-plan/reset",
	Method:     http.MethodPost,
	Handler:    xreq.Convert(ResetQuotaAction),
	Authorizer: iauth.FA(iauth.FeatureAPIKey, iauth.ActionUpdate),
}

type ResetQuotaReq struct {
	Quota *float64 `json:"quota,omitempty"`
}

type ResetQuotaResponse struct {
	ID            *string           `json:"id"`
	PreviousQuota *float64          `json:"previous_quota"`
	NewQuota      *float64          `json:"new_quota"`
	Balance       *ResetBalanceInfo `json:"balance"`
}

type ResetBalanceInfo struct {
	PreviousRemaining float64 `json:"previous_remaining"`
	NewRemaining      float64 `json:"new_remaining"`
	Used              float64 `json:"used"`
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

	var previousQuota *float64
	var previousRemaining float64

	// 获取重置前的配额信息
	plan, err := container.QuotaPlanManager.FetchQuotaPlan(req.Context(), &quota.QuotaPlanFilter{
		ID: apiKey.QuotaPlanID,
	})
	if err != nil {
		return nil, err
	}
	if plan != nil {
		previousQuota = plan.Quota
	}

	balance, err := container.QuotaPlanManager.FetchQuotaBalance(req.Context(), *apiKey.QuotaPlanID)
	if err != nil {
		return nil, err
	}
	if balance != nil && balance.Remaining != nil {
		previousRemaining = *balance.Remaining
	}

	// 重置配额余额（不更新 last_reset_at，避免影响定期重置调度）
	err = container.QuotaPlanManager.ResetBalance(req.Context(), *apiKey.QuotaPlanID, resetReq.Quota, false)
	if err != nil {
		return nil, err
	}

	// 重置 Redis 中的值为 quota 总量（RMB 配额按 1e8 定点整数存储）
	if apiKey.Key != nil && stateful.DefaultClientSet != nil && stateful.DefaultClientSet.RedisClient != nil {
		redisKey := stateful.AIUsedQuotaKey(*apiKey.Key)

		var resetQuota float64
		if resetReq.Quota != nil {
			resetQuota = *resetReq.Quota
		} else if plan != nil && plan.Quota != nil {
			resetQuota = *plan.Quota
		} else {
			return nil, xerror.WrapParamErrorWithMsg("quota is required")
		}

		unit := ""
		if plan != nil && plan.Unit != nil {
			unit = *plan.Unit
		}
		targetRedisValue := lib.QuotaToRedisValue(&resetQuota, &unit)

		currentValue, err := stateful.DefaultClientSet.RedisClient.GetInt64(redisKey)
		if err != nil {
			if !strings.Contains(err.Error(), "redigo: nil returned") {
				return nil, err
			}
			_, err = stateful.DefaultClientSet.RedisClient.IncrBy(redisKey, targetRedisValue)
		} else {
			delta := targetRedisValue - currentValue
			_, err = stateful.DefaultClientSet.RedisClient.IncrBy(redisKey, delta)
		}
		if err != nil {
			return nil, err
		}
	}

	// 获取更新后的配额计划信息
	newPlan, err := container.QuotaPlanManager.FetchQuotaPlan(req.Context(), &quota.QuotaPlanFilter{
		ID: apiKey.QuotaPlanID,
	})
	if err != nil {
		return nil, err
	}

	// 获取余额信息
	newBalance, err := container.QuotaPlanManager.FetchQuotaBalance(req.Context(), *apiKey.QuotaPlanID)
	if err != nil {
		return nil, err
	}

	newRemaining := float64(0)
	if newBalance != nil && newBalance.Remaining != nil {
		newRemaining = *newBalance.Remaining
	}

	newQuota := previousQuota
	if resetReq.Quota != nil {
		newQuota = resetReq.Quota
	} else if newPlan != nil {
		newQuota = newPlan.Quota
	}

	return &ResetQuotaResponse{
		ID:            oneReq.ID,
		PreviousQuota: previousQuota,
		NewQuota:      newQuota,
		Balance: &ResetBalanceInfo{
			PreviousRemaining: previousRemaining,
			NewRemaining:      newRemaining,
			Used:              0,
		},
	}, nil
}
