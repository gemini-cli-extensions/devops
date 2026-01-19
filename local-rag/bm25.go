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
	"io/ioutil"
	"math"
	"path/filepath"
	"sort"
	"strings"
)

// BM25 Constants
const (
	k1 = 1.2  // Term saturation parameter
	b  = 0.75 // Length normalization parameter
)

// Document represents a simple document with an ID and content
type Document struct {
	ID      int
	Content string
	Tokens  []string
}

// SearchResult holds the score and document ID
type SearchResult struct {
	DocID int
	Score float64
	Text  string
}

// BM25Index holds the index data structures
type BM25Index struct {
	Docs         []Document
	DocLengths   map[int]int            // Map of DocID -> Token Count
	TF           map[int]map[string]int // Map of DocID -> Term -> Frequency
	DF           map[string]int         // Map of Term -> Document Frequency
	AvgDocLength float64
	DocCount     int
}

// NewBM25Index initializes a new index
func NewBM25Index() *BM25Index {
	return &BM25Index{
		DocLengths: make(map[int]int),
		TF:         make(map[int]map[string]int),
		DF:         make(map[string]int),
		Docs:       make([]Document, 0),
	}
}

// AddDocument processes a document and adds it to the index
func (idx *BM25Index) AddDocument(id int, content string) {
	tokens := tokenize(content)
	docLen := len(tokens)

	// Store document metadata
	idx.Docs = append(idx.Docs, Document{ID: id, Content: content, Tokens: tokens})
	idx.DocLengths[id] = docLen
	idx.DocCount++

	// Calculate Term Frequencies for this document
	termCounts := make(map[string]int)
	for _, token := range tokens {
		termCounts[token]++
	}
	idx.TF[id] = termCounts

	// Update Document Frequencies (DF) - count unique terms per doc
	for term := range termCounts {
		idx.DF[term]++
	}

	// Update Average Document Length
	totalLen := 0
	for _, l := range idx.DocLengths {
		totalLen += l
	}
	idx.AvgDocLength = float64(totalLen) / float64(idx.DocCount)
}

// Search ranks documents based on the query using the BM25 formula
func (idx *BM25Index) Search(query string) []SearchResult {
	queryTerms := tokenize(query)
	scores := make(map[int]float64)

	for _, term := range queryTerms {
		// If term is not in our corpus, skip it
		df, exists := idx.DF[term]
		if !exists {
			continue
		}

		// Calculate IDF for this term
		// IDF = ln( (N - n(qi) + 0.5) / (n(qi) + 0.5) + 1 )
		idf := math.Log(1 + (float64(idx.DocCount)-float64(df)+0.5)/(float64(df)+0.5))

		// Score relevant documents
		for docID, termFreqs := range idx.TF {
			tf := float64(termFreqs[term])
			if tf == 0 {
				continue
			}

			docLen := float64(idx.DocLengths[docID])
			
			// Numerator: tf * (k1 + 1)
			numerator := tf * (k1 + 1)
			
			// Denominator: tf + k1 * (1 - b + b * (docLen / avgDocLen))
			denominator := tf + k1*(1-b+b*(docLen/idx.AvgDocLength))

			score := idf * (numerator / denominator)
			scores[docID] += score
		}
	}

	// Convert map to slice for sorting
	var results []SearchResult
	for docID, score := range scores {
		// Find the original text for display
		var text string
		for _, d := range idx.Docs {
			if d.ID == docID {
				text = d.Content
				break
			}
		}
		results = append(results, SearchResult{DocID: docID, Score: score, Text: text})
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// tokenize is a simple helper to lowercase and split text
// In a real app, use a stemmer (Snowball) and stop-word filter
func tokenize(text string) []string {
	text = strings.ToLower(text)
	// Remove punctuation (basic)
	f := func(c rune) bool {
		return c < 'a' || c > 'z' // keep only letters
	}
	// Split by non-letters
	return strings.FieldsFunc(text, f)
}

// loadFilesFromDirectory reads all files from a directory and adds them to the index
func loadFilesFromDirectory(idx *BM25Index, dirPath string, startID int) int {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("Error reading directory %s: %v\n", dirPath, err)
		return startID
	}

	docID := startID
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(dirPath, file.Name())
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", filePath, err)
			continue
		}

		idx.AddDocument(docID, string(content))
		fmt.Printf("Added document %d from %s\n", docID, file.Name())
		docID++
	}

	return docID
}

func main() {
	idx := NewBM25Index()

	// Load documents from patterns directory
	nextID := loadFilesFromDirectory(idx, "./patterns", 1)

	// Load documents from knowledge directory
	loadFilesFromDirectory(idx, "./knowledge", nextID)

	// 2. Perform Search
	query := "how to create a cloud build yaml"
	fmt.Printf("Search Query: '%s'\n", query)
	results := idx.Search(query)

	// 3. Display Results
	fmt.Println("---------------------------------------------------")
	fmt.Printf("%-5s | %-10s | %s\n", "Rank", "Score", "Content")
	fmt.Println("---------------------------------------------------")
	for i, res := range results {
		fmt.Printf("%-5d | %-10.4f | %s\n", i+1, res.Score, res.Text)
	}
}
