// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package ai_cache

import (
	"context"
	"time"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/shared"
)

// Defaults applied by the control plane when an optional rule field is omitted.
// They mirror the ai_cache_rules table column defaults (see db_ddl.sql).
const (
	DefaultCacheKeyStrategy = "lastQuestion"
	DefaultCacheTTL         = 0
	DefaultMaxBodyBytes     = 1048576
	DefaultMaxValueBytes    = 1048576
)

// AICacheRuleParam is the storage-level AI cache rule. The internal ID is the
// ordering/priority key and is never exposed through the Open API.
type AICacheRuleParam struct {
	ID               *int64
	Name             *string
	Cond             *string
	CacheKeyStrategy *string
	CacheTTL         *int
	MaxBodyBytes     *int64
	MaxValueBytes    *int64
	CreatedAt        *time.Time
	UpdatedAt        *time.Time
}

// AICacheStorager defines storage operations for the AI cache rule collection.
// ReplaceAll is executed inside a transaction provided by itxn.TxnStorager.
type AICacheStorager interface {
	// FetchAll returns all rules ordered by id ascending.
	FetchAll(ctx context.Context) ([]*AICacheRuleParam, error)
	// ReplaceAll atomically replaces the whole rule set with the given rules
	// (delete-all + insert-all, in slice order).
	ReplaceAll(ctx context.Context, rules []*AICacheRuleParam) error
}

// aiCacheRuleParamFromShared converts an API-level rule into the storage-level
// rule, filling the documented defaults for omitted optional fields so that
// every inserted row carries the full column set.
func aiCacheRuleParamFromShared(param *shared.AICacheRuleParam) *AICacheRuleParam {
	if param == nil {
		return nil
	}

	rule := &AICacheRuleParam{
		Name:             param.Name,
		Cond:             param.Cond,
		CacheKeyStrategy: param.CacheKeyStrategy,
		CacheTTL:         param.CacheTTL,
		MaxBodyBytes:     param.MaxBodyBytes,
		MaxValueBytes:    param.MaxValueBytes,
		CreatedAt:        param.CreatedAt,
		UpdatedAt:        param.UpdatedAt,
	}

	if rule.CacheKeyStrategy == nil {
		rule.CacheKeyStrategy = lib.PString(DefaultCacheKeyStrategy)
	}
	if rule.CacheTTL == nil {
		rule.CacheTTL = lib.PInt(DefaultCacheTTL)
	}
	if rule.MaxBodyBytes == nil {
		rule.MaxBodyBytes = lib.PInt64(DefaultMaxBodyBytes)
	}
	if rule.MaxValueBytes == nil {
		rule.MaxValueBytes = lib.PInt64(DefaultMaxValueBytes)
	}

	return rule
}

// aiCacheRuleParamToShared converts a storage-level rule into the API-level rule.
func aiCacheRuleParamToShared(rule *AICacheRuleParam) *shared.AICacheRuleParam {
	if rule == nil {
		return nil
	}

	return &shared.AICacheRuleParam{
		Name:             rule.Name,
		Cond:             rule.Cond,
		CacheKeyStrategy: rule.CacheKeyStrategy,
		CacheTTL:         rule.CacheTTL,
		MaxBodyBytes:     rule.MaxBodyBytes,
		MaxValueBytes:    rule.MaxValueBytes,
		CreatedAt:        rule.CreatedAt,
		UpdatedAt:        rule.UpdatedAt,
	}
}
