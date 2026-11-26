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

package router

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

func TestRouterDefaultSelectsAllRetrievers(t *testing.T) {
	ctx := context.Background()
	r, err := NewRetriever(ctx, &Config{
		Retrievers: map[string]retriever.Retriever{
			"one": &mockRetriever{},
			"two": &mockRetriever{},
		},
		FusionFunc: func(_ context.Context, result map[string][]*schema.Document) ([]*schema.Document, error) {
			require.Len(t, result, 2)
			require.Equal(t, []*schema.Document{{ID: "3"}}, result["one"])
			require.Equal(t, []*schema.Document{{ID: "3"}}, result["two"])
			return result["one"], nil
		},
	})
	require.NoError(t, err)
	docs, err := r.Retrieve(ctx, "3")
	require.NoError(t, err)
	require.Equal(t, []*schema.Document{{ID: "3"}}, docs)
}

func TestRouterDefaultWithSingleRetriever(t *testing.T) {
	ctx := context.Background()
	r, err := NewRetriever(ctx, &Config{
		Retrievers: map[string]retriever.Retriever{"only": &mockRetriever{}},
	})
	require.NoError(t, err)
	docs, err := r.Retrieve(ctx, "25")
	require.NoError(t, err)
	require.Equal(t, []*schema.Document{{ID: "2"}, {ID: "5"}}, docs)
}
