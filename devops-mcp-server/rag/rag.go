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
	"context"
	"devops-mcp-server/auth"
	"fmt"
	"log"
	"os"
	"sync"

	chromem "github.com/philippgille/chromem-go"
)

var initOnce sync.Once

// ListResult defines a generic struct to wrap a list of items.
type ListResult[T any] struct {
	Items []T `json:"items"`
}

type RagData struct {
	DB        *chromem.DB
	Pattern   *chromem.Collection
	Knowledge *chromem.Collection
}

var RagDB RagData

// loadRAG performs the one-time initialization.
func loadRAG(ctx context.Context) error {
	dbFile := os.Getenv("RAG_DB_PATH")
	RagDB = RagData{DB: chromem.NewDB()}
	if len(dbFile) < 1 {
		// RETURN AN ERROR
		return fmt.Errorf("Env variable RAG_DB_PATH is not set for RAG data file")
	}

	//check if file exists
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		// RETURN AN ERROR
		return fmt.Errorf("RAG_DB_PATH file does not exist, skipping import: %v", dbFile)
	}

	err := RagDB.DB.ImportFromFile(dbFile, "")
	if err != nil {
		log.Printf("Unable to import from the RAG DB file:%s - %v", dbFile, err)
		// This seems non-fatal based on your log, so we continue.
	}
	log.Printf("IMPORTED from the RAG DB file:%s - %v", dbFile, len(RagDB.DB.ListCollections()))

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

	RagDB.Pattern, err = RagDB.DB.GetOrCreateCollection("pattern", nil, vertexEmbeddingFunc)
	if err != nil {
		// RETURN AN ERROR
		return fmt.Errorf("Unable to get collection pattern: %w", err)
	}
	log.Printf("LOADED collection pattern: %v", RagDB.Pattern.Count())

	RagDB.Knowledge, err = RagDB.DB.GetOrCreateCollection("knowledge", nil, vertexEmbeddingFunc)
	if err != nil {
		// RETURN AN ERROR
		return fmt.Errorf("Unable to get collection knowledge: %w", err)
	}
	log.Printf("LOADED collection knowledge: %v", RagDB.Knowledge.Count())

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

func QueryKnowledge(ctx context.Context, query string) (*ListResult[chromem.Result], error) {
	ragDB, err := GetRAG(ctx)
	if err != nil {
		// Initialization failed, return the error.
		return nil, fmt.Errorf("RAG system not initialized: %w", err)
	}
	if ragDB.Knowledge == nil {
		return nil, fmt.Errorf("RagDB does not contain Knowledge collection, among %d collections", len(RagDB.DB.ListCollections()))
	}
	result, err := RagDB.Knowledge.Query(ctx, query, 3, nil, nil)
	if err != nil {
		log.Fatalf("Unable to Query collection pattern: %v", err)
	}
	return &ListResult[chromem.Result]{Items: result}, nil
}

func QueryPattern(ctx context.Context, query string) (*ListResult[chromem.Result], error) {
	ragDB, err := GetRAG(ctx)
	if err != nil {
		// Initialization failed, return the error.
		return nil, fmt.Errorf("RAG system not initialized: %w", err)
	}
	if ragDB.Pattern == nil {
		return nil, fmt.Errorf("RagDB does not contain Pattern collection, among %d collections", len(RagDB.DB.ListCollections()))
	}
	result, err := RagDB.Pattern.Query(ctx, query, 2, nil, nil)
	if err != nil {
		log.Fatalf("Unable to Query collection pattern: %v", err)
	}
	return &ListResult[chromem.Result]{Items: result}, nil
}
