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
	"path/filepath"

	chromem "github.com/philippgille/chromem-go"
)

func dbFile() string {
	dbFile := os.Getenv("RAG_DB_PATH")
	if len(dbFile) == 0 {
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		dbFile = filepath.Join(pwd, "devops-rag.db")
	}
	return dbFile
}

func setupRAGDB(ctx context.Context) (*chromem.DB, chromem.EmbeddingFunc, error) {
	vertexEmbeddingFunc, err := newGeminiEmbeddingFunc(ctx)
	if err != nil {
		return nil, nil, err
	}
	db := chromem.NewDB()
	dbFile := dbFile()
	if len(dbFile) > 0 {
		if _, err := os.Stat(dbFile); os.IsNotExist(err) {
			log.Printf("RAG_DB_PATH file does not exist, skipping import: %v", dbFile)
		} else {
			err := db.ImportFromFile(dbFile, "")
			log.Printf("Imported RAG with collections:%d", len(db.ListCollections()))
			if err != nil {
				return nil, nil, err
			}
		}
	}
	_, err = db.GetOrCreateCollection("knowledge", nil, vertexEmbeddingFunc)
	if err != nil {
		return nil, nil, err
	}
	_, err = db.GetOrCreateCollection("pattern", nil, vertexEmbeddingFunc)
	if err != nil {
		return nil, nil, err
	}
	return db, vertexEmbeddingFunc, nil
}

func BuildRagDatabase() {
	ctx := context.Background()

	db, embeddingFunc, err := setupRAGDB(ctx)
	if err != nil {
		log.Fatalf("Failed to setup RAG DB: %v", err)
	}

	// Upload local directories
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	collectionPattern := db.GetCollection("pattern", embeddingFunc)
	patternsDir := filepath.Join(pwd, "patterns")
	addDirectoryToRag(ctx, collectionPattern, patternsDir)

	collectionKnowledge := db.GetCollection("knowledge", embeddingFunc)
	knowledgeDir := filepath.Join(pwd, "knowledge")
	addDirectoryToRag(ctx, collectionKnowledge, knowledgeDir)

	documentSourceDir := filepath.Join(pwd, ".document-sources")
	if _, err := os.Stat(documentSourceDir); os.IsNotExist(err) {
		log.Printf("Dir does not exist: %v", documentSourceDir)
	} else {
		// Upload all files in the source directory to RAG
		addDirectoryToRag(ctx, collectionKnowledge, documentSourceDir)
	}

	// Export the database to a file
	dbFile := dbFile()
	if len(dbFile) > 0 {
		log.Printf("Exporting database Knowledge base docs:%d, Pattern docs:%d",
			collectionKnowledge.Count(),
			collectionPattern.Count())
		err = db.ExportToFile(dbFile, true, "")
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Database exported to %s", dbFile)
	}
}
