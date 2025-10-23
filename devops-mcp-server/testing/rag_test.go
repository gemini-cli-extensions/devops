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

// Cloud Build end-to-end tests
package testing

import (
	"context"
	"devops-mcp-server/auth"
	"devops-mcp-server/rag"
	"log"
	"testing"
)

func TestRAGQuery(t *testing.T) {
	ctx := context.Background()
	creds, err := auth.GetAuthToken(ctx)
	if creds.Token == "" || creds.ProjectId == "" || err != nil {
		t.Skipf("Skipping test! Google Cloud account not found %v : %v", creds, err)
	}

	patternResult, err := rag.QueryKnowledge(ctx, "Simple pipeline that deploys to Cloud Run")
	if err != nil {
		log.Fatalf("Unable to Query collection pattern: %v", err)
	}
	if (patternResult == nil) || (len(patternResult.Items) < 2) || (patternResult.Items[0].Content == "") {
		log.Fatalf("Failed to find knowledge: %v", patternResult)
	}

	knowledgeResult, err := rag.QueryKnowledge(ctx, "Package a Python application")
	if err != nil {
		log.Fatalf("Unable to Query collection knowledge: %v", err)
	}
	if (knowledgeResult == nil) || (len(knowledgeResult.Items) < 3) || (knowledgeResult.Items[0].Content == "") {
		log.Fatalf("Failed to find knowledge: %v", knowledgeResult)
	}

}
