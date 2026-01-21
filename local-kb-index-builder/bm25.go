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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"devops/lib/bm25"
)

// loadFilesFromDirectory reads all files from a directory (recursively) and adds them to the index
func loadFilesFromDirectory(idx *bm25.BM25Index, dirPath string, startID int) int {
	docID := startID
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", path, err)
			return nil
		}

		idx.AddDocument(docID, string(content), map[string]string{"source": path})
		fmt.Printf("Added document %d from %s\n", docID, path)
		docID++
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory %s: %v\n", dirPath, err)
	}

	return docID
}

func BuildBM25GobFiles() {
	// 1. Build Patterns Index
	patternsIdx := bm25.NewBM25Index()
	loadFilesFromDirectory(patternsIdx, "./patterns", 1)
	patternsFile := "patterns_index.gob"
	if err := bm25.SaveIndex(patternsFile, patternsIdx); err != nil {
		fmt.Printf("Error saving patterns index: %v\n", err)
	} else {
		fmt.Printf("Patterns index saved to %s\n", patternsFile)
	}

	// 2. Build Knowledge Index
	knowledgeIdx := bm25.NewBM25Index()
	nextID := loadFilesFromDirectory(knowledgeIdx, "./knowledge", 1)
	if _, err := os.Stat("./.document-sources"); err == nil {
		loadFilesFromDirectory(knowledgeIdx, "./.document-sources", nextID)
	}
	knowledgeFile := "knowledge_index.gob"
	if err := bm25.SaveIndex(knowledgeFile, knowledgeIdx); err != nil {
		fmt.Printf("Error saving knowledge index: %v\n", err)
	} else {
		fmt.Printf("Knowledge index saved to %s\n", knowledgeFile)
	}

	// 3. Perform a test search on knowledge
	query := "how to create a cloud build yaml"
	fmt.Printf("\nTest Search Query: '%s'\n", query)
	results := knowledgeIdx.Search(query, 3)

	fmt.Println("---------------------------------------------------")
	fmt.Printf("% -5s | % -10s | %s\n", "Rank", "Score", "Content")
	fmt.Println("---------------------------------------------------")
	for i, res := range results {
		content := res.Text
		if len(content) > 100 {
			content = content[:97] + "..."
		}
		fmt.Printf("% -5d | % -10.4f | %s\n", i+1, res.Score, strings.ReplaceAll(content, "\n", " "))
	}
}
