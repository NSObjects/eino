/*
 * Copyright 2026 CloudWeGo Authors
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

package internal

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func TestConcatStringClosesSource(t *testing.T) {
	wantErr := errors.New("stream failed")
	for _, tc := range []struct {
		name      string
		terminal  error
		converted bool
	}{
		{name: "source EOF", terminal: io.EOF},
		{name: "source error", terminal: wantErr},
		{name: "conversion error", terminal: wantErr, converted: true},
		{name: "conversion EOF", terminal: io.EOF, converted: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source, writer := schema.Pipe[string](3)
			source.SetAutomaticClose()
			t.Cleanup(source.Close)
			defer writer.Close()
			writer.Send("first", nil)

			reader := source
			if tc.converted {
				writer.Send("terminal", nil)
				reader = schema.StreamReaderWithConvert(source, func(chunk string) (string, error) {
					if chunk == "terminal" {
						return "", tc.terminal
					}
					return chunk, nil
				})
			} else {
				writer.Send("", tc.terminal)
			}

			result, err := concatString(reader)
			if tc.terminal == io.EOF {
				if err != nil || result != "first" {
					t.Fatalf("concatString() = %q, %v; want first, nil", result, err)
				}
			} else if !errors.Is(err, wantErr) {
				t.Fatalf("concatString() error = %v; want %v", err, wantErr)
			}

			if !writer.Send("unconsumed", nil) {
				t.Error("concatString did not close its source stream")
			}
		})
	}
}


// Exercise the middleware ownership boundary with a normal, non-automatic
// reader. Its Close is not idempotent, so this also detects a double close.
func TestToolResultOffloadingClosesConsumedSource(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("source failure")
	for _, tc := range []struct {
		name     string
		terminal error
	}{
		{name: "EOF", terminal: io.EOF},
		{name: "error", terminal: wantErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source, writer := schema.Pipe[string](3)
			defer writer.Close()
			writer.Send("first", nil)
			writer.Send("", tc.terminal)

			middleware := newToolResultOffloading(ctx, &toolResultOffloadingConfig{
				Backend:    newMockBackend(),
				TokenLimit: 1000,
			})
			wrapped := middleware.Streamable(func(context.Context, *compose.ToolInput) (*compose.StreamToolOutput, error) {
				return &compose.StreamToolOutput{Result: source}, nil
			})
			out, err := wrapped(ctx, &compose.ToolInput{Name: "test_tool", CallID: "call"})
			if !writer.Send("unconsumed", nil) {
				source.Close()
				t.Fatal("middleware did not close its consumed source")
			}
			if tc.terminal != io.EOF {
				if !errors.Is(err, wantErr) || out != nil {
					t.Fatalf("expected terminal error without replacement stream, got %v, %v", out, err)
				}
				return
			}
			if err != nil || out == nil || out.Result == nil {
				t.Fatalf("expected replacement stream, got %v, %v", out, err)
			}
			defer out.Result.Close()
			chunk, err := out.Result.Recv()
			if err != nil || chunk != "first" {
				t.Fatalf("replacement chunk = %q, %v", chunk, err)
			}
			if _, err := out.Result.Recv(); !errors.Is(err, io.EOF) {
				t.Fatalf("expected replacement EOF, got %v", err)
			}
		})
	}
}
