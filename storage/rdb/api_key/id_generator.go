// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
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

package api_key

import (
	"context"
	"fmt"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/api_key"
	"github.com/rainway-ai-gateway/ai-gateway-api/storage/rdb/internal/dao"
)

// RDBAPIKeyIDGenerator generates api-key-{seq} IDs using the database sequence
// table api_key_id_seq. It is safe for concurrent use across processes.
type RDBAPIKeyIDGenerator struct {
	dbCtxFactory lib.DBContextFactory
}

var _ api_key.APIKeyIDGenerator = (*RDBAPIKeyIDGenerator)(nil)

// NewRDBAPIKeyIDGenerator creates a new generator backed by the RDB sequence table.
func NewRDBAPIKeyIDGenerator(dbCtxFactory lib.DBContextFactory) *RDBAPIKeyIDGenerator {
	return &RDBAPIKeyIDGenerator{dbCtxFactory: dbCtxFactory}
}

// Generate allocates the next sequence number from the database and returns an
// ID in the form "api-key-{seq}".
func (g *RDBAPIKeyIDGenerator) Generate(ctx context.Context, productName string) (string, error) {
	dbCtx, err := g.dbCtxFactory(ctx)
	if err != nil {
		return "", err
	}

	seq, err := dao.TAPIKeyIDSeqAllocate(dbCtx, productName)
	if err != nil {
		return "", fmt.Errorf("allocate api-key id sequence: %w", err)
	}

	return fmt.Sprintf("api-key-%d", seq), nil
}
