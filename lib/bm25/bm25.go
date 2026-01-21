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

package bm25

import (
	"bytes"
	"encoding/gob"
	"math"
	"os"
	"sort"
	"strings"
)

// BM25 Constants
const (
	K1 = 1.2  // Term saturation parameter
	B  = 0.75 // Length normalization parameter
)

// Document represents a simple document with an ID and content
type Document struct {
	ID       int               `json:"id"`
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Tokens   []string          `json:"-"`
}

// SearchResult holds the score and document ID
type SearchResult struct {
	DocID    int               `json:"doc_id"`
	Score    float64           `json:"relevance_score"`
	Text     string            `json:"content"`
	Metadata map[string]string `json:"metadata,omitempty"`
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
func (idx *BM25Index) AddDocument(id int, content string, metadata map[string]string) {
	tokens := Tokenize(content)
	docLen := len(tokens)

	// Store document metadata
	idx.Docs = append(idx.Docs, Document{ID: id, Content: content, Tokens: tokens, Metadata: metadata})
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
func (idx *BM25Index) Search(query string, limit int) []SearchResult {
	queryTerms := Tokenize(query)
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

			// Numerator: tf * (K1 + 1)
			numerator := tf * (K1 + 1)

			// Denominator: tf + K1 * (1 - B + B * (docLen / avgDocLen))
			denominator := tf + K1*(1-B+B*(docLen/idx.AvgDocLength))

			score := idf * (numerator / denominator)
			scores[docID] += score
		}
	}

	// Convert map to slice for sorting
	var results []SearchResult
	for docID, score := range scores {
		// Find the original text for display
		var text string
		var metadata map[string]string
		for _, d := range idx.Docs {
			if d.ID == docID {
				text = d.Content
				metadata = d.Metadata
				break
			}
		}
		results = append(results, SearchResult{DocID: docID, Score: score, Text: text, Metadata: metadata})
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results
}

// Tokenize is a simple helper to lowercase and split text
// In a real app, use a stemmer (Snowball) and stop-word filter
func Tokenize(text string) []string {
	text = strings.ToLower(text)
	// Remove punctuation (basic)
	f := func(c rune) bool {
		return c < 'a' || c > 'z' // keep only letters
	}
	// Split by non-letters
	return strings.FieldsFunc(text, f)
}

// SaveIndex serializes the BM25Index to a binary file using Gob.
func SaveIndex(filename string, idx *BM25Index) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	return encoder.Encode(idx)
}

// LoadIndex deserializes the BM25Index from a binary file using Gob.
func LoadIndex(filename string) (*BM25Index, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var idx BM25Index
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&idx); err != nil {
		return nil, err
	}
	return &idx, nil
}

// LoadIndexFromBytes deserializes the BM25Index from a byte slice using Gob.
func LoadIndexFromBytes(data []byte) (*BM25Index, error) {
	var idx BM25Index
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&idx); err != nil {
		return nil, err
	}
	return &idx, nil
}
