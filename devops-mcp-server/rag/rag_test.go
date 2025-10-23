package rag

import (
	"os"
	"testing"
)

func TestRAGInit(t *testing.T) {

	if RagDB.DB == nil {
		t.Fatalf("init() failed to load: %v", os.Getenv("RAG_DB_PATH"))
	}
}
