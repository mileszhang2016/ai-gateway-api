package subcluster

import (
	"net/http"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

type GlobalListParam struct {
	PoolName string `form:"pool_name"`
}

// deprecated, endpoint registration removed per optimization plan v1.2
// var GlobalListEndpoint = &xreq.Endpoint{
// 	Path:       "/global-sub-clusters",
// 	Method:     http.MethodGet,
// 	Handler:    xreq.Convert(GlobalListAction),
// 	Authorizer: iauth.FA(iauth.FeatureProduct, iauth.ActionRead),
// }

var _ xreq.Handler = GlobalListAction

func GlobalListAction(req *http.Request) (interface{}, error) {
	return globalListActionProcess(req)
}

func globalListActionProcess(req *http.Request) ([]*OneData, error) {
	var param GlobalListParam
	if err := xreq.BindForm(req, &param); err != nil {
		return nil, xerror.WrapParamError(err)
	}

	subClusterList, err := container.SubClusterManager.SubClusterList(req.Context(), &icluster_conf.SubClusterFilter{})
	if err != nil {
		return nil, err
	}

	if len(subClusterList) == 0 {
		return nil, nil
	}

	list := make([]*OneData, 0)
	for _, one := range subClusterList {
		if param.PoolName != "" && (one.InstancePool == nil || one.InstancePool.Name != param.PoolName) {
			continue
		}
		list = append(list, newOneData(one))
	}
	return list, nil
}
