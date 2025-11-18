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

package ragclient

import (
	"bytes"
	"context"
	"devops-mcp-server/auth"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"

	chromem "github.com/philippgille/chromem-go"
)

//go:embed devops-rag.db
var embeddedDB []byte

type RagClientImpl struct {
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

type RagClient interface {
	Queryknowledge(ctx context.Context, query string) (string, error)
	QueryPatterns(ctx context.Context, query string) (string, error)
}

// loadRAG performs the one-time initialization.
func loadRAG(ctx context.Context) (RagClient, error) {
	ragClient := &RagClientImpl{DB: chromem.NewDB()}
	reader := bytes.NewReader(embeddedDB)
	err := ragClient.DB.ImportFromReader(reader, "")
	if err != nil {
		log.Printf("Unable to import from the RAG DB file: %v", err)
		return nil, err
	}
	log.Printf("IMPORTED from the RAG DB collections: %v", len(ragClient.DB.ListCollections()))

	creds, err := auth.GetAuthToken(ctx)
	if err != nil {
		log.Printf("Error: Google Cloud account is required: %v", err)
		// RETURN AN ERROR
		return nil, fmt.Errorf("Google Cloud account is required: %w", err)
	}

	vertexEmbeddingFunc := chromem.NewEmbeddingFuncVertex(
		creds.Token,
		creds.ProjectId,
		chromem.EmbeddingModelVertexEnglishV4)
	ragClient.Knowledge, err = ragClient.DB.GetOrCreateCollection("knowledge", nil, vertexEmbeddingFunc)
	if err != nil {
		return nil, fmt.Errorf("Unable to get collection knowledge: %w", err)
	}
	log.Printf("LOADED collection knowledge: %v", ragClient.Knowledge.Count())
	ragClient.Pattern, err = ragClient.DB.GetOrCreateCollection("pattern", nil, vertexEmbeddingFunc)
	if err != nil {
		return nil, fmt.Errorf("Unable to get collection pattern: %w", err)
	}
	log.Printf("LOADED collection pattern: %v", ragClient.Pattern.Count())

	log.Print("RAG Init Completed!")
	return ragClient, nil // Success
}

// contextKey is a private type to use as a key for context values.
type contextKey string

const (
	ragClientKey contextKey = "ragClient"
)

// ClientFrom returns the RagClient stored in the context, if any.
func ClientFrom(ctx context.Context) (RagClient, bool) {
	client, ok := ctx.Value(ragClientKey).(RagClient)
	return client, ok
}

// ContextWithClient returns a new context with the provided RagClient.
func ContextWithClient(ctx context.Context, client RagClient) context.Context {
	return context.WithValue(ctx, ragClientKey, client)
}

// NewClient creates a new Client.
func NewClient(ctx context.Context) (RagClient, error) {
	return loadRAG(ctx)
}


func (r *RagClientImpl) QueryPatterns(ctx context.Context, query string) (string, error) {
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

func (r *RagClientImpl) Queryknowledge(ctx context.Context, query string) (string, error) {
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
