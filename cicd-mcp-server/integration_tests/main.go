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
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cicd-mcp-server/artifactregistry"
	"cicd-mcp-server/cloudrun"
	"cicd-mcp-server/osv"

	artifactregistryclient "cicd-mcp-server/artifactregistry/client"
	cloudrunclient "cicd-mcp-server/cloudrun/client"
	iamclient "cicd-mcp-server/iam/client"
	osvclient "cicd-mcp-server/osv/client"

	mcpclient "github.com/mark3labs/mcp-go/client"
	mcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx := context.Background()

	// Create the server
	server, arClient, crClient, osvClient := createMCPServer(ctx)

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
	// Artifact Registry Tests
	testSetupRepository(ctx, arClient)
	// Cloud Run Tests
	testListServices(ctx, crClient)
	testDeployToCloudRunFromImage(ctx, crClient)            // Tests the deploy_cloudrun_service_from_image tool with a new service.
	testDeployToCloudRunFromImageNewRevision(ctx, crClient) // Tests the deploy_cloudrun_service_from_image tool with a preexisting service.
	testDeployToCloudRunFromSource(ctx, crClient)
	// OSV Tests
	testScanSecrets(ctx, osvClient)
	testScanSecretsWithSecret(ctx, osvClient)
}

func createMCPServer(ctx context.Context) (*mcpserver.Server, artifactregistryclient.ArtifactRegistryClient, cloudrunclient.CloudRunClient, osvclient.OsvClient) {
	server := mcpserver.NewServer(&mcpserver.Implementation{Name: "cicd-mcp-server"}, nil)

	arClient, err := artifactregistryclient.NewArtifactRegistryClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Artifact Registry client: %v", err)
	}
	iamClient, err := iamclient.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create IAM client: %v", err)
	}
	crClient, err := cloudrunclient.NewCloudRunClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Cloud Run client: %v", err)
	}
	osvClient, err := osvclient.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create OSV client: %v", err)
	}

	arHandler := &artifactregistry.Handler{
		ArClient:  arClient,
		IamClient: iamClient,
	}
	arHandler.Register(server)

	crHandler := &cloudrun.Handler{
		CrClient: crClient,
	}
	crHandler.Register(server)

	osvHandler := &osv.Handler{
		OsvClient: osvClient,
	}
	osvHandler.Register(server)

	return server, arClient, crClient, osvClient
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
	req.Params.Name = "create_artifact_repository"
	req.Params.Arguments = args

	log.Println("Calling tool 'create_artifact_repository'...")

	resp, err := mcpClient.CallTool(ctx, req)
	if err != nil {
		log.Fatalf("Tool call failed: %v", err)
	}

	if resp.IsError {
		log.Fatalf("Tool returned an error: %v", resp.Content)
	}

	log.Println("Tool call successful.")

	// Clean up the repository
	defer func() {
		log.Println("Cleaning up repository...")
		err := arClient.DeleteRepository(ctx, projectID, "us-central1", "integration-test-repo")
		if err != nil {
			log.Printf("Failed to delete repository: %v", err)
		}
	}()

	// Verify that the repository was created
	log.Println("Verifying repository creation...")
	_, err = arClient.GetRepository(ctx, projectID, "us-central1", "integration-test-repo")
	if err != nil {
		log.Fatalf("Failed to get repository after creation: %v", err)
	}

	log.Println("Repository verification successful.")
}

func testListServices(ctx context.Context, crClient cloudrunclient.CloudRunClient) {
	log.Println("--- Running test: ListServices ---")
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

	serviceNames := []string{
		"integration-test-service-1",
		"integration-test-service-2",
	}

	// Create services for the test
	for _, serviceName := range serviceNames {
		_, err = crClient.CreateService(ctx, projectID, "us-central1", serviceName, "us-docker.pkg.dev/cloudrun/container/hello", 8080)
		if err != nil {
			log.Fatalf("Failed to create service: %v", err)
		}
	}

	// Clean up the services
	defer func() {
		log.Println("Cleaning up services...")
		for _, serviceName := range serviceNames {
			err := crClient.DeleteService(ctx, projectID, "us-central1", serviceName)
			if err != nil {
				log.Printf("Failed to delete service: %v", err)
			}
		}
	}()

	args := map[string]any{
		"project_id": projectID,
		"location":   "us-central1",
	}

	var req mcp.CallToolRequest
	req.Params.Name = "list_cloudrun_services"
	req.Params.Arguments = args

	log.Println("Calling tool 'list_cloudrun_services'...")

	resp, err := mcpClient.CallTool(ctx, req)
	if err != nil {
		log.Fatalf("Tool call failed: %v", err)
	}

	if resp.IsError {
		log.Fatalf("Tool returned an error: %v", resp.Content)
	}

	log.Println("Tool call successful.")

	// Extract output from resp and verify services were listed
	contentMap, ok := resp.StructuredContent.(map[string]interface{})
	if !ok {
		log.Fatalf("StructuredContent was not a map. Got: %T", resp.StructuredContent)
	}
	services, ok := contentMap["services"].([]interface{})
	if !ok {
		log.Fatalf("Content map did not contain a 'services' key with a list of services. Got: %T", contentMap["services"])
	}

	gotServiceList := make(map[string]string)
	for _, item := range services {
		serviceMap, ok := item.(map[string]interface{})
		if !ok {
			log.Fatalf("An item in the services list was not a map. Got: %T", item)
		}

		name, ok := serviceMap["name"].(string)
		if !ok {
			log.Fatalf("Service name is not a string. Got: %T", serviceMap["name"])
		}

		parts := strings.Split(name, "/")
		serviceName := parts[len(parts)-1]
		gotServiceList[serviceName] = ""
	}

	for _, serviceName := range serviceNames {
		if _, ok := gotServiceList[serviceName]; !ok {
			log.Fatalf("Expected service %q was not found in the response.", serviceName)
		}

	}

	log.Println("Services verification successful.")
}

// Tests the deploy_cloudrun_service_from_image tool with a new private service.
func testDeployToCloudRunFromImage(ctx context.Context, crClient cloudrunclient.CloudRunClient) {
	log.Println("--- Running test: CreateService ---")
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
		"service_name":  "integration-test-service",
		"revision_name": "integration-test-service-revision-v1",
		"image_url":     "us-docker.pkg.dev/cloudrun/container/hello",
		"port":          8080,
	}

	var req mcp.CallToolRequest
	req.Params.Name = "deploy_cloudrun_service_from_image"
	req.Params.Arguments = args

	log.Println("Calling tool 'deploy_cloudrun_service_from_image'...")

	resp, err := mcpClient.CallTool(ctx, req)
	if err != nil {
		log.Fatalf("Tool call failed: %v", err)
	}

	if resp.IsError {
		log.Fatalf("Tool returned an error: %v", resp.Content)
	}

	log.Println("Tool call successful.")

	// Clean up the service
	defer func() {
		log.Println("Cleaning up service...")
		err := crClient.DeleteService(ctx, projectID, "us-central1", "integration-test-service")
		if err != nil {
			log.Printf("Failed to delete service: %v", err)
		}
	}()

	// Verify that the service was created
	log.Println("Verifying service creation...")
	_, err = crClient.GetService(ctx, projectID, "us-central1", "integration-test-service")
	if err != nil {
		log.Fatalf("Failed to get service after creation: %v", err)
	}

	log.Println("Service verification successful.")
}

// Tests the deploy_cloudrun_service_from_image tool with a preexisting service (private service).
func testDeployToCloudRunFromImageNewRevision(ctx context.Context, crClient cloudrunclient.CloudRunClient) {
	log.Println("--- Running test: CreateServiceRevision ---")
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

	serviceName := "integration-test-service-revision"

	// Create a preexisting service for the test
	_, err = crClient.CreateService(ctx, projectID, "us-central1", serviceName, "us-docker.pkg.dev/cloudrun/container/hello", 8080)
	if err != nil {
		log.Fatalf("Failed to create service: %v", err)
	}

	// Clean up the service
	defer func() {
		log.Println("Cleaning up service...")
		err := crClient.DeleteService(ctx, projectID, "us-central1", serviceName)
		if err != nil {
			log.Printf("Failed to delete service: %v", err)
		}
	}()

	args := map[string]any{
		"project_id":    projectID,
		"location":      "us-central1",
		"service_name":  serviceName,
		"image_url":     "us-docker.pkg.dev/cloudrun/container/hello",
		"revision_name": fmt.Sprintf("%s-v2", serviceName),
		"port":          8080,
	}

	var req mcp.CallToolRequest
	req.Params.Name = "deploy_cloudrun_service_from_image"
	req.Params.Arguments = args

	log.Println("Calling tool 'deploy_cloudrun_service_from_image'...")

	resp, err := mcpClient.CallTool(ctx, req)
	if err != nil {
		log.Fatalf("Tool call failed: %v", err)
	}

	if resp.IsError {
		log.Fatalf("Tool returned an error: %v", resp.Content)
	}

	log.Println("Tool call successful.")

	// Verify that the revision was created
	log.Println("Verifying revision creation...")
	service, err := crClient.GetService(ctx, projectID, "us-central1", serviceName)
	if err != nil {
		log.Fatalf("Failed to get service after creation: %v", err)
	}
	revision, err := crClient.GetRevision(ctx, service)
	if err != nil {
		log.Fatalf("Failed to get revision after creation: %v", err)
	}
	wantRevision := fmt.Sprintf("projects/%s/locations/%s/services/%s/revisions/%s", projectID, args["location"], args["service_name"], args["revision_name"])
	if revision.Name != wantRevision {
		log.Fatalf("Revision name mismatch: expected 'integration-test-service-v2', got '%s'", revision.Name)
	}

	log.Println("Revision verification successful.")
}

// Tests creating a private cloudrun service.
func testDeployToCloudRunFromSource(ctx context.Context, crClient cloudrunclient.CloudRunClient) {
	log.Println("--- Running test: DeployToCloudRunFromSource ---")
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

	// Create a temporary directory with a simple main.go file
	tmpDir, err := os.MkdirTemp("", "test-source-*")
	if err != nil {
		log.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mainGoContent := `package main
import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	log.Print("starting server...")
	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request from %s", r.RemoteAddr)
	fmt.Fprintf(w, "Hello, World!")
}
`
	if err := os.WriteFile(tmpDir+"/main.go", []byte(mainGoContent), 0644); err != nil {
		log.Fatalf("Failed to write main.go file: %v", err)
	}

	args := map[string]any{
		"project_id":   projectID,
		"location":     "us-central1",
		"service_name": "integration-test-service-from-source",
		"source":       tmpDir,
		"port":         8080,
	}

	var req mcp.CallToolRequest
	req.Params.Name = "deploy_cloudrun_service_from_source"
	req.Params.Arguments = args

	log.Println("Calling tool 'deploy_cloudrun_service_from_source'...")

	resp, err := mcpClient.CallTool(ctx, req)
	if err != nil {
		log.Fatalf("Tool call failed: %v", err)
	}

	if resp.IsError {
		log.Fatalf("Tool returned an error: %v", resp.Content)
	}

	log.Println("Tool call successful.")

	// Clean up the service
	defer func() {
		log.Println("Cleaning up service...")
		err := crClient.DeleteService(ctx, projectID, "us-central1", "integration-test-service-from-source")
		if err != nil {
			log.Printf("Failed to delete service: %v", err)
		}
	}()

	// Verify that the service was created
	log.Println("Verifying service creation...")
	_, err = crClient.GetService(ctx, projectID, "us-central1", "integration-test-service-from-source")
	if err != nil {
		log.Fatalf("Failed to get service after creation: %v", err)
	}

	log.Println("Service verification successful.")
}

func testScanSecrets(ctx context.Context, oClient osvclient.OsvClient) {
	log.Println("--- Running test: ScanSecrets ---")
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

	// Create a temporary directory with a dummy file
	tmpDir, err := os.MkdirTemp("", "test-secrets-*")
	if err != nil {
		log.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dummyFile := filepath.Join(tmpDir, "test-file.txt")
	if err := os.WriteFile(dummyFile, []byte("no secrets here"), 0644); err != nil {
		log.Fatalf("Failed to write dummy file: %v", err)
	}

	args := map[string]any{
		"root":               tmpDir,
		"ignore_directories": []string{},
	}

	var req mcp.CallToolRequest
	req.Params.Name = "scan_code_for_secrets"
	req.Params.Arguments = args

	log.Println("Calling tool 'scan_code_for_secrets'...")

	resp, err := mcpClient.CallTool(ctx, req)
	if err != nil {
		log.Fatalf("Tool call failed: %v", err)
	}

	if resp.IsError {
		log.Fatalf("Tool returned an error: %v", resp.Content)
	}

	log.Println("Tool call successful.")
}

func testScanSecretsWithSecret(ctx context.Context, oClient osvclient.OsvClient) {
	log.Println("--- Running test: ScanSecretsWithSecret ---")
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

	// Create a temporary directory with a file containing a fake secret
	tmpDir, err := os.MkdirTemp("", "test-secrets-*")
	if err != nil {
		log.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	secretFile := filepath.Join(tmpDir, "vulnerable_config.env")
	fakeSecret := "AIzaSyC_TestTokenForScannerDetection123" // Key needs to start with AIza and be 39 characters long.
	if err := os.WriteFile(secretFile, []byte(fmt.Sprintf("GOOGLE_API_KEY=%s", fakeSecret)), 0644); err != nil {
		log.Fatalf("Failed to write secret file: %v", err)
	}

	args := map[string]any{
		"root":               tmpDir,
		"ignore_directories": []string{},
	}

	var req mcp.CallToolRequest
	req.Params.Name = "scan_code_for_secrets"
	req.Params.Arguments = args

	log.Println("Calling tool 'scan_code_for_secrets'...")

	resp, err := mcpClient.CallTool(ctx, req)
	if err != nil {
		log.Fatalf("Tool call failed: %v", err)
	}

	if resp.IsError {
		log.Fatalf("Tool returned an error: %v", resp.Content)
	}

	log.Println("Tool call successful.")

	// Verify that the fake secret was found in the report
	contentMap, ok := resp.StructuredContent.(map[string]interface{})
	if !ok {
		log.Fatalf("StructuredContent was not a map. Got: %T", resp.StructuredContent)
	}
	report, ok := contentMap["report"].(string)
	if !ok {
		log.Fatalf("Content map did not contain a 'report' key with a string report. Got: %T", contentMap["report"])
	}

	if !strings.Contains(report, fakeSecret) {
		log.Fatalf("Expected fake secret %q not found in scan report: %s", fakeSecret, report)
	}

	log.Println("Secret scan verification successful.")
}
