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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudwego/eino/schema"
)

func TestNewRetrieverRejectsInvalidQueryLimit(t *testing.T) {
	ctx := context.Background()
	r, err := NewRetriever(ctx, nil)
	require.ErrorContains(t, err, "config is required")
	require.Nil(t, r)

	r, err = NewRetriever(ctx, &Config{
		OrigRetriever: &mockRetriever{},
		RewriteHandler: func(context.Context, string) ([]string, error) {
			return []string{"1"}, nil
		},
		MaxQueriesNum: -1,
	})
	require.ErrorContains(t, err, "MaxQueriesNum must not be negative")
	require.Nil(t, r)
}

func TestMultiQueryPositiveAndDefaultLimits(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name  string
		limit int
		want  []*schema.Document
	}{
		{"default", 0, []*schema.Document{{ID: "1"}, {ID: "2"}, {ID: "3"}, {ID: "4"}, {ID: "5"}}},
		{"one", 1, []*schema.Document{{ID: "1"}}},
		{"two", 2, []*schema.Document{{ID: "1"}, {ID: "2"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewRetriever(ctx, &Config{
				OrigRetriever: &mockRetriever{},
				RewriteHandler: func(context.Context, string) ([]string, error) {
					return []string{"1", "2", "3", "4", "5", "6"}, nil
				},
				MaxQueriesNum: tc.limit,
			})
			require.NoError(t, err)
			docs, err := r.Retrieve(ctx, "question")
			require.NoError(t, err)
			require.Equal(t, tc.want, docs)
		})
	}
}
