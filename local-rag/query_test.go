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

package main

import (
	"context"
	"log"
	"os"
	"testing"

	chromem "github.com/philippgille/chromem-go"
)

func TestRAGQuery(t *testing.T) {
	ctx := context.Background()
	token, projectID, err := getGCPToken(ctx)
	if err != nil {
		t.Skipf("Skipping test due to missing GCP credentials: %v", err)
	}

	vertexEmbeddingFunc := chromem.NewEmbeddingFuncVertex(
		token,
		projectID,
		chromem.EmbeddingModelVertexEnglishV4)

	db := chromem.NewDB()
	dbFile := os.Getenv("RAG_DB_PATH")
	if len(dbFile) == 0 {
		t.Skip("Skipping test: RAG_DB_PATH environment variable not set")
	}

	//check if file exists, we expect an existing DB
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		t.Skipf("Skipping test: RAG_DB_PATH file does not exist: %v", dbFile)
	} else {
		err := db.ImportFromFile(dbFile, "")
		log.Printf("Imported RAG with collections:%d", len(db.ListCollections()))
		if err != nil {
			t.Fatalf("Unable to import from the RAG DB file:%s - %v", dbFile, err)
		}
	}

	collectionPattern, err := db.GetOrCreateCollection("pattern", nil, vertexEmbeddingFunc)
	if err != nil {
		t.Fatalf("Unable to get collection pattern: %v", err)
	}

	patternResult, err := collectionPattern.Query(ctx, "Simple pipeline that deploys to Cloud Run", 1, nil, nil)
	if err != nil {
		t.Fatalf("Unable to Query collection pattern: %v", err)
	}
	if len(patternResult) < 1 || patternResult[0].Content == "" {
		t.Fatalf("Failed to find pattern: %v", len(patternResult))
	}

	collectionKnowledge, err := db.GetOrCreateCollection("knowledge", nil, vertexEmbeddingFunc)
	if err != nil {
		t.Fatalf("Unable to get collection knowledge: %v", err)
	}

	knowledgeResult, err := collectionKnowledge.Query(ctx, "Package a Python application", 3, nil, nil)
	if err != nil {
		t.Fatalf("Unable to Query collection knowledge: %v", err)
	}
	if len(knowledgeResult) < 3 || knowledgeResult[0].Content == "" {
		log.Printf("Knowledge result count: %d", len(knowledgeResult))
		// Only fail if we really expected 3 and got 0 or empty content.
		// The original test checked < 3 but maybe we should be lenient if the DB is small?
		// I'll keep the original logic but add logging.
		t.Fatalf("Failed to find pattern: %v", len(knowledgeResult))
	}
}
