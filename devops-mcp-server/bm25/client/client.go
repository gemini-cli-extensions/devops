// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bm25

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"devops/lib/bm25"
)

//go:embed patterns_index.gob
var patternsIndexData []byte

//go:embed knowledge_index.gob
var knowledgeIndexData []byte

type BM25Client interface {
	Queryknowledge(ctx context.Context, query string) (string, error)
	QueryPatterns(ctx context.Context, query string) (string, error)
}

// NewClient creates a new Client.
func NewClient(ctx context.Context) (BM25Client, error) {
	return loadDoc(ctx)
}

func loadDoc(ctx context.Context) (BM25Client, error) {
	bm25Client := &BM25ClientImpl{}

	patternsIdx, err := bm25.LoadIndexFromBytes(patternsIndexData)
	if err != nil {
		return nil, fmt.Errorf("failed to load patterns index: %w", err)
	}
	bm25Client.Patterns = patternsIdx

	knowledgeIdx, err := bm25.LoadIndexFromBytes(knowledgeIndexData)
	if err != nil {
		return nil, fmt.Errorf("failed to load knowledge index: %w", err)
	}
	bm25Client.Knowledge = knowledgeIdx

	return bm25Client, nil
}

type BM25ClientImpl struct {
	Patterns  *bm25.BM25Index
	Knowledge *bm25.BM25Index
}

func (b *BM25ClientImpl) Queryknowledge(ctx context.Context, query string) (string, error) {
	results := b.Knowledge.Search(query, 3)

	// Marshal to JSON
	jsonData, err := json.Marshal(results)
	if err != nil {
		return "", fmt.Errorf("failed to marshal results: %w", err)
	}
	return string(jsonData), nil
}

func (b *BM25ClientImpl) QueryPatterns(ctx context.Context, query string) (string, error) {
	results := b.Patterns.Search(query, 3)

	// Marshal to JSON
	jsonData, err := json.Marshal(results)
	if err != nil {
		return "", fmt.Errorf("failed to marshal results: %w", err)
	}
	return string(jsonData), nil
}
