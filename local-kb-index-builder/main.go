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
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: local-rag <command>")
		fmt.Println("Commands:")
		fmt.Println("  download - Download all required files to .document-sources folder")
		fmt.Println("  rag      - Create the RAG database from downloaded files and local docs that can be used later on")
		fmt.Println("  bm25     - Create a BM 25 index gob that can be used later on")
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "download":
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		documentSourceDir := filepath.Join(pwd, ".document-sources")
		DownloadDocuments(documentSourceDir)
	case "rag":
		BuildRagDatabase()
	case "bm25":
		BuildBM25GobFiles()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
