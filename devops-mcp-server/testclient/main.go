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
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	ctx := context.Background()
	const serverURL = "http://localhost:8080"

	// 1. Create a new Streamable HTTP client
	mcpClient, err := client.NewStreamableHttpClient(serverURL, nil)
	if err != nil {
		log.Fatalf("Failed to create mcp-go HTTP client: %v", err)
	}

	// 2. Start the client's connection loop
	if err := mcpClient.Start(ctx); err != nil {
		log.Fatalf("Failed to start mcp-go client: %v", err)
	}
	defer mcpClient.Close()

	// 3. ***FIX***: Explicitly perform the MCP initialization handshake
	// This call is blocking and ensures the client is initialized.
	var initReq mcp.InitializeRequest
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION // Use the SDK's constant
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    "my-test-client",
		Version: "1.0.0",
	}

	if _, err := mcpClient.Initialize(ctx, initReq); err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}
	// The client is now guaranteed to be initialized.

	// 4. Define the arguments for the tool
	args := map[string]any{
		"project_id":    "ishamirulinda-sdlc",
		"location":      "us-central1",
		"repository_id": "my-test-repo1",
		"format":        "DOCKER",
	}

	// 5. Create the CallToolRequest
	var req mcp.CallToolRequest
	req.Params.Name = "artifactregistry.create_repository"
	req.Params.Arguments = args

	fmt.Println("Calling tool 'devops/artifactregistry.create_repository' using mark3labs client...")

	// 6. Call the tool
	resp, err := mcpClient.CallTool(ctx, req)
	if err != nil {
		log.Fatalf("Tool call failed: %v", err)
	}

	// 7. Check for an error in the response
	if resp.IsError {
		log.Fatalf("Tool returned an error: %v", resp.Content)
	}

	fmt.Println("Tool call successful.")

	// 8. Pretty-print the result
	resultJSON, err := json.MarshalIndent(resp.Content, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal result to JSON: %v", err)
	}

	fmt.Printf("Result:\n%s\n", string(resultJSON))
}
