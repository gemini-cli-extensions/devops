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

	chromem "github.com/philippgille/chromem-go"
)

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

func init() {

	ctx := context.Background()

	dbFile := os.Getenv("RAG_DB_PATH")
	RagDB = RagData{DB: chromem.NewDB()}
	if len(dbFile) < 1 {
		log.Fatalf("Env variable RAG_DB_PATH is not set for RAG data file %v", os.Getenv("RAG_DB_PATH"))
	}
	//check if file exists, we expect an existing DB
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		log.Fatalf("RAG_DB_PATH file does not exist, skipping import: %v", dbFile)
	} else {
		err := RagDB.DB.ImportFromFile(dbFile, "")
		if err != nil {
			log.Printf("Unable to import from the RAG DB file:%s - %v", dbFile, err)
		}
		log.Printf("IMPORTED from the RAG DB file:%s - %v", dbFile, len(RagDB.DB.ListCollections()))
	}

	creds, err := auth.GetAuthToken(ctx)
	if err != nil {
		log.Printf("Error: Google Cloud account is required: %v", err)
		//TODO: Should we fail MCP server startup if Google Cloud account is not setup?
		//For now if Google Cloud account is not setup, we will silently fail
		return
	}

	vertexEmbeddingFunc := chromem.NewEmbeddingFuncVertex(
		creds.Token,
		creds.ProjectId,
		chromem.EmbeddingModelVertexEnglishV4)

	RagDB.Pattern, err = RagDB.DB.GetOrCreateCollection("pattern", nil, vertexEmbeddingFunc)
	if err != nil {
		log.Fatalf("Unable to get collection pattern: %v", err)
	} else {
		log.Printf("LOADED collection pattern: %v", RagDB.Pattern.Count())
	}

	RagDB.Knowledge, err = RagDB.DB.GetOrCreateCollection("knowledge", nil, vertexEmbeddingFunc)
	if err != nil {
		log.Fatalf("Unable to get collection knowledge: %v", err)
	} else {
		log.Printf("LOADED collection knowledge: %v", RagDB.Knowledge.Count())
	}
	log.Print("Init Completed!")
}

func QueryKnowledge(ctx context.Context, query string) (*ListResult[chromem.Result], error) {
	if RagDB.Knowledge == nil {
		return nil, fmt.Errorf("RagDB does not contain Knowledge collection, among %d collections", len(RagDB.DB.ListCollections()))
	}
	result, err := RagDB.Knowledge.Query(ctx, query, 3, nil, nil)
	if err != nil {
		log.Fatalf("Unable to Query collection pattern: %v", err)
	}
	return &ListResult[chromem.Result]{Items: result}, nil
}

func QueryPattern(ctx context.Context, query string) (*ListResult[chromem.Result], error) {
	if RagDB.Pattern == nil {
		return nil, fmt.Errorf("RagDB does not contain Pattern collection, among %d collections", len(RagDB.DB.ListCollections()))
	}
	result, err := RagDB.Pattern.Query(ctx, query, 2, nil, nil)
	if err != nil {
		log.Fatalf("Unable to Query collection pattern: %v", err)
	}
	return &ListResult[chromem.Result]{Items: result}, nil
}
