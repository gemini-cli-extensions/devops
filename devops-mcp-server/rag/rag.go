// Copyright 2024 Google LLC
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

package rag

import (
	"bytes"
	"context"
	"devops-mcp-server/auth"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	chromem "github.com/philippgille/chromem-go"
)

var initOnce sync.Once

//go:embed devops-rag.db
var embeddedDB []byte

type RagData struct {
	DB        *chromem.DB
	Pattern   *chromem.Collection
	Knowledge *chromem.Collection
}

// Only expose what the LLM needs to read.
type Result struct {
	Content    string            `json:"content"`
	Metadata   map[string]string `json:"metadata,omitempty"` // Source info
	Similarity float32           `json:"relevance_score"`    // Helps LLM weigh confidence
}

var RagDB RagData

// loadRAG performs the one-time initialization.
func loadRAG(ctx context.Context) error {
	RagDB = RagData{DB: chromem.NewDB()}
	reader := bytes.NewReader(embeddedDB)
	err := RagDB.DB.ImportFromReader(reader, "")
	if err != nil {
		log.Printf("Unable to import from the RAG DB file: %v", err)
		return err
	}
	log.Printf("IMPORTED from the RAG DB collections: %v", len(RagDB.DB.ListCollections()))

	creds, err := auth.GetAuthToken(ctx)
	if err != nil {
		log.Printf("Error: Google Cloud account is required: %v", err)
		// RETURN AN ERROR
		return fmt.Errorf("Google Cloud account is required: %w", err)
	}

	vertexEmbeddingFunc := chromem.NewEmbeddingFuncVertex(
		creds.Token,
		creds.ProjectId,
		chromem.EmbeddingModelVertexEnglishV4)
	RagDB.Knowledge, err = RagDB.DB.GetOrCreateCollection("knowledge", nil, vertexEmbeddingFunc)
	if err != nil {
		return fmt.Errorf("Unable to get collection knowledge: %w", err)
	}
	log.Printf("LOADED collection knowledge: %v", RagDB.Pattern.Count())
	RagDB.Pattern, err = RagDB.DB.GetOrCreateCollection("pattern", nil, vertexEmbeddingFunc)
	if err != nil {
		return fmt.Errorf("Unable to get collection pattern: %w", err)
	}
	log.Printf("LOADED collection pattern: %v", RagDB.Pattern.Count())

	log.Print("RAG Init Completed!")
	return nil // Success
}

// GetRAG returns the initialized RAG database.
// It ensures initialization is performed exactly once.
func GetRAG(ctx context.Context) (RagData, error) {
	var initErr error
	initOnce.Do(func() {
		// Use a background context for the one-time init
		initErr = loadRAG(context.Background())
	})

	return RagDB, initErr
}

func (r *RagData) QueryPattern(ctx context.Context, query string) (string, error) {
	results, err := r.Pattern.Query(ctx, query, 2, nil, nil)
	if err != nil {
		log.Fatalf("Unable to Query collection pattern: %v", err)
	}
	cleanResults := make([]Result, len(results))
	for i, r := range results {
		cleanResults[i] = Result{
			Content:    r.Content,
			Metadata:   r.Metadata,
			Similarity: r.Similarity,
		}
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(cleanResults)
	if err != nil {
		return "", fmt.Errorf("failed to marshal results: %w", err)
	}
	return string(jsonData), nil
}

func (r *RagData) Queryknowledge(ctx context.Context, query string) (string, error) {
	results, err := r.Knowledge.Query(ctx, query, 2, nil, nil)
	if err != nil {
		log.Fatalf("Unable to Query collection knowledge: %v", err)
	}
	cleanResults := make([]Result, len(results))
	for i, r := range results {
		cleanResults[i] = Result{
			Content:    r.Content,
			Metadata:   r.Metadata,
			Similarity: r.Similarity,
		}
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(cleanResults)
	if err != nil {
		return "", fmt.Errorf("failed to marshal results: %w", err)
	}
	return string(jsonData), nil
}
