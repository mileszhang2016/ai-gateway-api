package product_pool

import (
	"context"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/ibasic"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful/container"
)

func getDefaultProduct(ctx context.Context) (*ibasic.Product, error) {
	name := stateful.DefaultConfig.RunTime.AIRouteInnerProductName
	products, err := container.ProductManager.FetchProducts(ctx, &ibasic.ProductFilter{Name: &name})
	if err != nil {
		return nil, err
	}
	if len(products) != 1 {
		return nil, xerror.WrapParamErrorWithMsg("Default Product Not Exist")
	}
	return products[0], nil
}