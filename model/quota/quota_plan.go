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

package quota

import "context"

// QuotaPlanParam 定义配额计划参数
type QuotaPlanParam struct {
	ID                    *int64  `json:"id"`
	Unlimited             *bool   `json:"unlimited"`
	PassWhenNoEnoughQuota *bool   `json:"pass_when_no_enough_quota"`
	Quota                 *int64  `json:"quota"`
	Unit                  *string `json:"unit"`
	ResetPeriod           *string `json:"reset_period"`
}

// QuotaPlanFilter 定义配额计划过滤条件
type QuotaPlanFilter struct {
	ID          *int64
	Unlimited   *bool
	ResetPeriod []string
}

// QuotaPlanStorager 定义配额计划存储接口
type QuotaPlanStorager interface {
	CreateQuotaPlan(ctx context.Context, param *QuotaPlanParam) (int64, error)
	FetchQuotaPlan(ctx context.Context, filter *QuotaPlanFilter) (*QuotaPlanParam, error)
	FetchQuotaPlanList(ctx context.Context, filter *QuotaPlanFilter) ([]*QuotaPlanParam, error)
	UpdateQuotaPlan(ctx context.Context, filter *QuotaPlanFilter, param *QuotaPlanParam) (int64, error)
	DeleteQuotaPlan(ctx context.Context, filter *QuotaPlanFilter) error
}
