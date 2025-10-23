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

package prompts

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestDesignPrompt(t *testing.T) {
	ctx := context.Background()

	// Create a minimal mcp.Server instance for testing.
	impl := &mcp.Implementation{
		Name:    "test-devops",
		Title:   "Test Google Cloud DevOps MCP Server",
		Version: "v0.0.1",
	}
	opts := &mcp.ServerOptions{
		Instructions: "Test Google Cloud DevOps MCP Server",
		HasResources: false,
	}
	server := mcp.NewServer(impl, opts)

	// Call DesignPrompt to register the prompt with the test server.
	DesignPrompt(ctx, server)

	// Connect a client to the server using in-memory transports.
	t1, t2 := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, t1, nil)
	if err != nil {
		log.Fatalf("failed to connect server: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.1"}, nil)
	clientSession, err := client.Connect(ctx, t2, nil)
	if err != nil {
		log.Fatalf("failed to connect client: %v", err)
	}

	// Verify that the prompt is registered.
	prompts, err := clientSession.ListPrompts(ctx, &mcp.ListPromptsParams{}) // Changed from &mcp.ListPromptsRequest{} to &mcp.ListPromptsParams{}
	if err != nil {
		log.Fatalf("failed to list prompts: %v", err)
	}

	foundPrompt := false
	var designPrompt *mcp.Prompt
	for _, p := range prompts.Prompts {
		if p.Name == "devops:design" {
			foundPrompt = true
			designPrompt = p
			break
		}
	}

	if !foundPrompt {
		t.Error("Prompt 'devops:design' not found on the server")
	}

	if designPrompt == nil {
		t.Fatal("designPrompt is nil")
	}

	expectedTitle := "Design and implement a Google Cloud based CI/CD pipeline."
	if designPrompt.Title != expectedTitle {
		t.Errorf("Prompt title mismatch: got %q, want %q", designPrompt.Title, expectedTitle)
	}

	if len(designPrompt.Arguments) != 1 {
		t.Fatalf("Expected 1 argument, got %d", len(designPrompt.Arguments))
	}

	expectedArgName := "query"
	if designPrompt.Arguments[0].Name != expectedArgName {
		t.Errorf("Prompt argument name mismatch: got %q, want %q", designPrompt.Arguments[0].Name, expectedArgName)
	}

	// Test the prompt handler's behavior.
	query := "create a simple CI/CD pipeline"
	getPromptParams := &mcp.GetPromptParams{
		Name: "devops:design", // Name is part of GetPromptParams for GetPrompt
		Arguments: map[string]string{"query": query},
	}

	result, err := clientSession.GetPrompt(ctx, getPromptParams)
	if err != nil {
		t.Fatalf("GetPrompt returned an error: %v", err)
	}

	if result.Description != "Helps design and implement GCP CI/CD pipelines." {
		t.Errorf("Handler description mismatch: got %q, want %q", result.Description, "Helps design and implement GCP CI/CD pipelines.")
	}

	if len(result.Messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(result.Messages))
	}

	if result.Messages[0].Role != "user" {
		t.Errorf("Message role mismatch: got %q, want \"user\"", result.Messages[0].Role)
	}

	textContent, ok := result.Messages[0].Content.(*mcp.TextContent)
	if !ok {
		t.Fatalf("Expected Content to be *mcp.TextContent, got %T", result.Messages[0].Content)
	}
	expectedContent := fmt.Sprintf(promptCICDText, query)
	if textContent.Text != expectedContent {
		t.Errorf("Message content mismatch: got %q, want %q", textContent.Text, expectedContent)
	}

	// Shut down the session.
	if err := clientSession.Close(); err != nil {
		log.Fatal(err)
	}
	if err := serverSession.Wait(); err != nil {
		log.Fatal(err)
	}
}