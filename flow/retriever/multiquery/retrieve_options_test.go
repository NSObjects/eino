/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package multiquery

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

type multiQueryOptionFixture struct {
	Tenant string
}

type multiQueryOptionRetriever struct{}

func (*multiQueryOptionRetriever) Retrieve(_ context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	common := retriever.GetCommonOptions(nil, opts...)
	specific := retriever.GetImplSpecificOptions[multiQueryOptionFixture](nil, opts...)
	if common.Index == nil || *common.Index != "tenant-index" ||
		common.TopK == nil || *common.TopK != 2 ||
		common.DSLInfo["tenant"] != "tenant-a" ||
		specific.Tenant != "tenant-a" {
		return nil, fmt.Errorf("query %q did not receive all caller options", query)
	}
	return []*schema.Document{{ID: query}}, nil
}

func TestMultiQueryForwardsOptionsToEveryRetrieval(t *testing.T) {
	ctx := context.Background()
	r, err := NewRetriever(ctx, &Config{
		OrigRetriever: &multiQueryOptionRetriever{},
		RewriteHandler: func(context.Context, string) ([]string, error) {
			return []string{"first", "second", "third"}, nil
		},
	})
	require.NoError(t, err)
	docs, err := r.Retrieve(ctx, "question",
		retriever.WithIndex("tenant-index"),
		retriever.WithTopK(2),
		retriever.WithDSLInfo(map[string]any{"tenant": "tenant-a"}),
		retriever.WrapImplSpecificOptFn(func(o *multiQueryOptionFixture) {
			o.Tenant = "tenant-a"
		}),
	)
	require.NoError(t, err)
	require.Equal(t, []*schema.Document{{ID: "first"}, {ID: "second"}, {ID: "third"}}, docs)
}
