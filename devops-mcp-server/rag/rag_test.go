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
