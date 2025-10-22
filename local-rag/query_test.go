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

	"cloud.google.com/go/auth"
	"cloud.google.com/go/auth/credentials"
	chromem "github.com/philippgille/chromem-go"
)

func gcpAuthHelper(ctx context.Context, t *testing.T) (tokenValue, projectID string) {
	// Use Application Default Credentials to get a TokenSource
	scopes := []string{"https://www.googleapis.com/auth/cloud-platform"}
	creds, err := credentials.DetectDefault(&credentials.DetectOptions{
		Scopes: scopes,
	})
	if err != nil {
		log.Fatalf("Failed to find default credentials: %v", err)
	}

	projectID, err = creds.ProjectID(ctx)
	if err != nil {
		log.Fatalf("Failed to get project ID: %v", err)
	}
	if projectID == "" {
		//Try quota project
		projectID, err = creds.QuotaProjectID(ctx)
		if err != nil {
			log.Fatalf("Failed to get project ID: %v", err)
		}
		if projectID == "" {
			log.Fatalf(`
			No Project ID found in Application Default Credentials. 
			This can happen if credentials are user-based or the project hasn't been explicitly set 
			e.g., via gcloud auth application-default set-quota-project.
			Error:%v`, err)
		}
	}
	// We need an access token
	var token *auth.Token
	token, err = creds.TokenProvider.Token(ctx)
	if err != nil {
		log.Fatalf("Failed to retrieve access token: %v", err)
	}

	return token.Value, projectID
}

func TestRAGQuery(t *testing.T) {
	ctx := context.Background()
	token, projectID := gcpAuthHelper(ctx, t)

	vertexEmbeddingFunc := chromem.NewEmbeddingFuncVertex(
		token,
		projectID,
		chromem.EmbeddingModelVertexEnglishV4)

	db := chromem.NewDB()
	dbFile := os.Getenv("RAG_DB_PATH")
	if len(dbFile) > 0 {

	}
	//check if file exists, we expect an existing DB
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		log.Fatalf("RAG_DB_PATH file does not exist, skipping import: %v", dbFile)
	} else {
		err := db.ImportFromFile(dbFile, "")
		log.Printf("Imported RAG with collections:%d", len(db.ListCollections()))
		if err != nil {
			log.Fatalf("Unable to import from the RAG DB file:%s - %v", dbFile, err)
		}
	}

	collectionPattern, err := db.GetOrCreateCollection("pattern", nil, vertexEmbeddingFunc)
	if err != nil {
		log.Fatalf("Unable to get collection pattern: %v", err)
	}

	patternResult, err := collectionPattern.Query(ctx, "Simple pipeline that deploys to Cloud Run", 1, nil, nil)
	if err != nil {
		log.Fatalf("Unable to Query collection pattern: %v", err)
	}
	if len(patternResult) < 1 || patternResult[0].Content == "" {
		log.Fatalf("Failed to find pattern: %v", len(patternResult))
	}

	collectionKnowledge, err := db.GetOrCreateCollection("knowledge", nil, vertexEmbeddingFunc)
	if err != nil {
		log.Fatalf("Unable to get collection knowledge: %v", err)
	}

	knowledgeResult, err := collectionKnowledge.Query(ctx, "Package a Python application", 3, nil, nil)
	if err != nil {
		log.Fatalf("Unable to Query collection knowledge: %v", err)
	}
	if len(knowledgeResult) < 3 || knowledgeResult[0].Content == "" {
		log.Fatalf("Failed to find pattern: %v", len(knowledgeResult))
	}
}
