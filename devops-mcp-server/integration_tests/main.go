package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"devops-mcp-server/artifactregistry"

	artifactregistryclient "devops-mcp-server/artifactregistry/client"
	iamclient "devops-mcp-server/iam/client"

	mcpclient "github.com/mark3labs/mcp-go/client"
	mcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx := context.Background()

	// Create the server
	server, arClient := createMCPServer(ctx)

	// Start the server in a goroutine
	go func() {
		log.Println("Starting server...")
		handler := mcpserver.NewStreamableHTTPHandler(func(*http.Request) *mcpserver.Server {
			return server
		}, nil)
		if err := http.ListenAndServe(":8080", handler); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for the server to start
	time.Sleep(2 * time.Second)

	// Run the tests
	testSetupRepository(ctx, arClient)
}

func createMCPServer(ctx context.Context) (*mcpserver.Server, artifactregistryclient.ArtifactRegistryClient) {
	server := mcpserver.NewServer(&mcpserver.Implementation{Name: "devops-mcp-server"}, nil)

	arClient, err := artifactregistryclient.NewArtifactRegistryClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Artifact Registry client: %v", err)
	}
	iamClient, err := iamclient.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create IAM client: %v", err)
	}

	ctxWithDeps := artifactregistryclient.ContextWithClient(ctx, arClient)
	ctxWithDeps = iamclient.ContextWithClient(ctxWithDeps, iamClient)

	if err := artifactregistry.AddTools(ctxWithDeps, server); err != nil {
		log.Fatalf("Failed to add artifactregistry tools: %v", err)
	}

	return server, arClient
}

func testSetupRepository(ctx context.Context, arClient artifactregistryclient.ArtifactRegistryClient) {
	log.Println("--- Running test: SetupRepository ---")
	const serverURL = "http://localhost:8080"

	mcpClient, err := mcpclient.NewStreamableHttpClient(serverURL, nil)
	if err != nil {
		log.Fatalf("Failed to create mcp-go HTTP client: %v", err)
	}

	if err := mcpClient.Start(ctx); err != nil {
		log.Fatalf("Failed to start mcp-go client: %v", err)
	}
	defer mcpClient.Close()

	var initReq mcp.InitializeRequest
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    "integration-test-client",
		Version: "1.0.0",
	}

	if _, err := mcpClient.Initialize(ctx, initReq); err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		log.Fatal("GCP_PROJECT_ID environment variable not set")
	}

	args := map[string]any{
		"project_id":    projectID,
		"location":      "us-central1",
		"repository_id": "integration-test-repo",
		"format":        "DOCKER",
	}

	var req mcp.CallToolRequest
	req.Params.Name = "artifactregistry.setup_repository"
	req.Params.Arguments = args

	fmt.Println("Calling tool 'artifactregistry.setup_repository'...")

	resp, err := mcpClient.CallTool(ctx, req)
	if err != nil {
		log.Fatalf("Tool call failed: %v", err)
	}

	if resp.IsError {
		log.Fatalf("Tool returned an error: %v", resp.Content)
	}

	fmt.Println("Tool call successful.")

	// Verify that the repository was created
	log.Println("Verifying repository creation...")
	_, err = arClient.GetRepository(ctx, projectID, "us-central1", "integration-test-repo")
	if err != nil {
		log.Fatalf("Failed to get repository after creation: %v", err)
	}

	log.Println("Repository verification successful.")
}
