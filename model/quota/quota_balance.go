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

import (
	"context"
	"time"
)

// QuotaBalanceParam 定义配额余额参数
type QuotaBalanceParam struct {
	ID          *int64     `json:"id"`
	QuotaPlanID *int64     `json:"quota_plan_id"`
	Used        *int64     `json:"used"`
	Remaining   *int64     `json:"remaining"`
	LastResetAt *time.Time `json:"last_reset_at"`
}

// QuotaBalanceFilter 定义配额余额过滤条件
type QuotaBalanceFilter struct {
	ID          *int64
	QuotaPlanID *int64
}

// QuotaBalanceStorager 定义配额余额存储接口
type QuotaBalanceStorager interface {
	CreateQuotaBalance(ctx context.Context, param *QuotaBalanceParam) (int64, error)
	FetchQuotaBalance(ctx context.Context, filter *QuotaBalanceFilter) (*QuotaBalanceParam, error)
	FetchQuotaBalanceList(ctx context.Context, filter *QuotaBalanceFilter) ([]*QuotaBalanceParam, error)
	UpdateQuotaBalance(ctx context.Context, filter *QuotaBalanceFilter, param *QuotaBalanceParam) (int64, error)
	DeleteQuotaBalance(ctx context.Context, filter *QuotaBalanceFilter) error
}
