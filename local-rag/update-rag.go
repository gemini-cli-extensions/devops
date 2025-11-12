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
	"runtime"
	"strconv"

	chromem "github.com/philippgille/chromem-go"
	"github.com/tmc/langchaingo/textsplitter"
)

func addDirectoryToRag(ctx context.Context, collection *chromem.Collection, dir string) {
	var docs []chromem.Document
	log.Printf("Uploading directory %s to collection: %v", dir, collection.Name)
	splitter := textsplitter.NewMarkdownTextSplitter(
		textsplitter.WithChunkSize(1000),
		textsplitter.WithChunkOverlap(150),
	)

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
					Content:  chunk,
					Metadata: map[string]string{"source": path},
				}
				docs = append(docs, doc)
			}
		}
		return nil
	})

	if len(docs) > 0 {
		threads := 5
		if threads > runtime.NumCPU() {
			threads = runtime.NumCPU()
		}
		err := collection.AddDocuments(context.Background(), docs, threads)
		if err != nil {
			log.Printf("Error adding documents from %s to collection: %v", dir, err)
		}
	}
}
