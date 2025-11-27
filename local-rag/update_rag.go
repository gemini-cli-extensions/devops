// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
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
	"strconv"
	"strings"

	chromem "github.com/philippgille/chromem-go"
	"github.com/tmc/langchaingo/textsplitter"
)

func addDirectoryToRag(ctx context.Context, collection *chromem.Collection, dir string) {
	var docs []chromem.Document
	log.Printf("Uploading directory %s to collection: %v", dir, collection.Name)
	//For embedding models Gemini limits to 2048 tokens.
	//Assuming 4 charact per token and ~15% overlap
	//set chunk size to max possible values, any larger and we hit the limit
	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(5000),
		textsplitter.WithChunkOverlap(750),
	)
	sourcePath := ""
	dirPath := strings.Split(dir, "/")
	if len(dirPath) > 1 {
		//sourcePath = strings.Join(dirPath[len(dirPath)-1], "/")
		sourcePath =  dirPath[len(dirPath)-1]
	}
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			_, err := collection.GetByID(ctx, path)
			if err == nil {
				// log.Printf("Doc found %s: %v", path, err)
				// Skip if doc is already loaded
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				log.Printf("Error reading file %s: %v", path, err)
				return nil
			}
			//split contents to chunks
			chunks, err := splitter.SplitText(string(content))
			if err != nil {
				log.Printf("Error chunking file %s: %v", path, err)
				return nil
			}
			for index, chunk := range chunks {
				chunkId := path + "_" + strconv.Itoa(index)
				_, err := collection.GetByID(ctx, chunkId)
				if err == nil {
					// log.Printf("Doc found %s: %v", path, err)
					// Skip if doc is already loaded
					return nil
				}
				doc := chromem.Document{
					ID:       chunkId,
					Content:  string(chunk),
					Metadata: map[string]string{"source":  strings.ReplaceAll(path,dir,sourcePath)},
				}
				docs = append(docs, doc)
			}
		}
		return nil
	})

	if len(docs) > 0 {
		threads := 5

		batchSize := 100
		for i := 0; i < len(docs); i += batchSize {
			end := i + batchSize
			if end > len(docs) {
				end = len(docs)
			}

			batch := docs[i:end]
			log.Printf("Adding batch %d-%d (total %d) to collection: %s", i, end, len(docs), collection.Name)
			err := collection.AddDocuments(context.Background(), batch, threads)
			if err != nil {
				log.Printf("Error adding batch %d-%d from %s to collection: %v", i, end, dir, err)
			}
		}
	}
}
