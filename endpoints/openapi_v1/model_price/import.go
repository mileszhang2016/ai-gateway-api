// Copyright(c) 2026 The Infinity AI Gateway Authors.
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

package model_price

import (
	"net/http"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/imodel_price"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

// ImportEndpoint imports model prices from a model-list.yaml file.
var ImportEndpoint = &xreq.Endpoint{
	Path:       "/model-prices/import",
	Method:     http.MethodPost,
	Handler:    xreq.Convert(ImportAction),
	Authorizer: iauth.FA(iauth.FeatureModelPrice, iauth.ActionCreate),
}

type importResponse struct {
	ImportedCount int      `json:"imported_count"`
	SkippedCount  int      `json:"skipped_count"`
	Errors        []string `json:"errors"`
}

// ImportAction handles POST /model-prices/import.
func ImportAction(req *http.Request) (interface{}, error) {
	mode := req.FormValue("mode")
	if mode == "" {
		mode = string(imodel_price.ImportModeReplace)
	}

	file, _, err := req.FormFile("file")
	if err != nil {
		return nil, xerror.WrapParamErrorWithMsg("file is required")
	}
	defer file.Close()

	modelList, err := imodel_price.ParseModelListYAML(file)
	if err != nil {
		return nil, xerror.WrapParamErrorWithMsg("failed to parse yaml: %v", err)
	}

	if err := imodel_price.ValidateImportFile(modelList); err != nil {
		return nil, xerror.WrapParamErrorWithMsg("invalid import file: %v", err)
	}

	imported, skipped, err := container.ModelPriceManager.ImportModelPrices(req.Context(), modelList.Models, mode)
	if err != nil {
		return nil, err
	}

	return &importResponse{
		ImportedCount: imported,
		SkippedCount:  skipped,
		Errors:        []string{},
	}, nil
}
