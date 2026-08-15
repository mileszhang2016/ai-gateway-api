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
	"strconv"

	"github.com/gorilla/mux"

	"github.com/infinity-ai-gateway/ai-gateway-api/model/imodel_price"
)

func idFromURI(req *http.Request) *int64 {
	vars := mux.Vars(req)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		return nil
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil
	}
	return &id
}

func queryFilter(req *http.Request) *imodel_price.ModelPriceFilter {
	q := req.URL.Query()
	filter := &imodel_price.ModelPriceFilter{}

	if v := q.Get("provider"); v != "" {
		filter.Provider = &v
	}
	if v := q.Get("model"); v != "" {
		filter.Model = &v
	}
	if v := q.Get("mode"); v != "" {
		filter.Mode = &v
	}
	return filter
}

func pageFilter(req *http.Request) (page int, pageSize int) {
	q := req.URL.Query()
	page = 1
	pageSize = 50
	if v := q.Get("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			page = p
		}
	}
	if v := q.Get("page_size"); v != "" {
		if ps, err := strconv.Atoi(v); err == nil && ps > 0 {
			if ps > 1000 {
				ps = 1000
			}
			pageSize = ps
		}
	}
	return page, pageSize
}

func emptyString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
