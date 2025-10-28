package rag

import (
	"os"
	"testing"
)

func TestRAGInit(t *testing.T) {
	t.Skip("Skipping test due to missing RAG_DB_PATH env var")
	if RagDB.DB == nil {
		t.Fatalf("init() failed to load: %v", os.Getenv("RAG_DB_PATH"))
	}
}
