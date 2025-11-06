package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"devops-mcp-server/artifactregistry"
	"devops-mcp-server/cloudrun"
	"devops-mcp-server/cloudstorage"

	artifactregistryclient "devops-mcp-server/artifactregistry/client"
	cloudrunclient "devops-mcp-server/cloudrun/client"
	cloudstorageclient "devops-mcp-server/cloudstorage/client"
	iamclient "devops-mcp-server/iam/client"

	mcpclient "github.com/mark3labs/mcp-go/client"
	mcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx := context.Background()

	// Create the server
	server, arClient, csClient, crClient := createMCPServer(ctx)

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
	// Cloud Storage Tests
	testCreateBucket(ctx, csClient)
	testUploadFile(ctx, csClient)
	testUploadDirectory(ctx, csClient)
	// Cloud Run Tests
	testCreateService(ctx, crClient)
	testCreateRevision(ctx, crClient)
	testCreateServiceFromSource(ctx, crClient)
}

func createMCPServer(ctx context.Context) (*mcpserver.Server, artifactregistryclient.ArtifactRegistryClient, cloudstorageclient.CloudStorageClient, cloudrunclient.CloudRunClient) {
	server := mcpserver.NewServer(&mcpserver.Implementation{Name: "devops-mcp-server"}, nil)

	arClient, err := artifactregistryclient.NewArtifactRegistryClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Artifact Registry client: %v", err)
	}
	iamClient, err := iamclient.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create IAM client: %v", err)
	}
	csClient, err := cloudstorageclient.NewCloudStorageClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Cloud Storage client: %v", err)
	}
	crClient, err := cloudrunclient.NewCloudRunClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Cloud Run client: %v", err)
	}

	ctxWithDeps := artifactregistryclient.ContextWithClient(ctx, arClient)
	ctxWithDeps = iamclient.ContextWithClient(ctxWithDeps, iamClient)
	ctxWithDeps = cloudstorageclient.ContextWithClient(ctxWithDeps, csClient)
	ctxWithDeps = cloudrunclient.ContextWithClient(ctxWithDeps, crClient)

	if err := artifactregistry.AddTools(ctxWithDeps, server); err != nil {
		log.Fatalf("Failed to add artifactregistry tools: %v", err)
	}
	if err := cloudstorage.AddTools(ctxWithDeps, server); err != nil {
		log.Fatalf("Failed to add cloudstorage tools: %v", err)
	}
	if err := cloudrun.AddTools(ctxWithDeps, server); err != nil {
		log.Fatalf("Failed to add cloudrun tools: %v", err)
	}

	return server, arClient, csClient, crClient
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

	log.Println("Calling tool 'artifactregistry.setup_repository'...")

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

func testCreateBucket(ctx context.Context, csClient cloudstorageclient.CloudStorageClient) {
	log.Println("--- Running test: CreateBucket ---")
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

	bucketName := fmt.Sprintf("%s-integration-test-bucket", projectID)

	args := map[string]any{
		"project_id":  projectID,
		"bucket_name": bucketName,
	}

	var req mcp.CallToolRequest
	req.Params.Name = "cloudstorage.create_bucket"
	req.Params.Arguments = args

	log.Println("Calling tool 'cloudstorage.create_bucket'...")

	resp, err := mcpClient.CallTool(ctx, req)
	if err != nil {
		log.Fatalf("Tool call failed: %v", err)
	}

	if resp.IsError {
		log.Fatalf("Tool returned an error: %v", resp.Content)
	}

	log.Println("Tool call successful.")

	// Clean up the bucket
	defer func() {
		log.Println("Cleaning up bucket...")
		err := csClient.DeleteBucket(ctx, bucketName)
		if err != nil {
			log.Printf("Failed to delete bucket: %v", err)
		}
	}()

	// Verify that the bucket was created
	log.Println("Verifying bucket creation...")
	err = csClient.CheckBucketExists(ctx, bucketName)
	if err != nil {
		log.Fatalf("Failed to get bucket after creation: %v", err)
	}

	log.Println("Bucket verification successful.")
}

func testUploadFile(ctx context.Context, csClient cloudstorageclient.CloudStorageClient) {
	log.Println("--- Running test: UploadFile ---")
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

	bucketName := fmt.Sprintf("%s-integration-test-bucket-upload-file", projectID)
	objectName := "test-object"

	// Create a bucket for the test
	err = csClient.CreateBucket(ctx, projectID, bucketName)
	if err != nil {
		log.Fatalf("Failed to create bucket: %v", err)
	}

	// Clean up the bucket
	defer func() {
		log.Println("Cleaning up bucket...")
		err = csClient.DeleteBucket(ctx, bucketName)
		if err != nil {
			log.Printf("Failed to delete bucket: %v", err)
		}
	}()

	tmpFile, err := os.CreateTemp("", "test-file-*.txt")
	if err != nil {
		log.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte("hello world")); err != nil {
		log.Fatalf("Failed to write to temporary file: %v", err)
	}
	tmpFile.Close()

	args := map[string]any{
		"bucket_name": bucketName,
		"object_name": objectName,
		"file_path":   tmpFile.Name(),
	}

	var req mcp.CallToolRequest
	req.Params.Name = "cloudstorage.upload_file"
	req.Params.Arguments = args

	log.Println("Calling tool 'cloudstorage.upload_file'...")

	resp, err := mcpClient.CallTool(ctx, req)
	if err != nil {
		log.Fatalf("Tool call failed: %v", err)
	}

	if resp.IsError {
		log.Fatalf("Tool returned an error: %v", resp.Content)
	}

	log.Println("Tool call successful.")

	// Clean up the file
	defer func() {
		log.Println("Cleaning up file...")
		err := csClient.DeleteObject(ctx, bucketName, objectName)
		if err != nil {
			log.Printf("Failed to delete object: %v", err)
		}
	}()

	// Verify that the file was uploaded
	log.Println("Verifying file upload...")
	err = csClient.CheckObjectExists(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("Failed to get object after upload: %v", err)
	}

	log.Println("File upload verification successful.")
}

func testUploadDirectory(ctx context.Context, csClient cloudstorageclient.CloudStorageClient) {
	log.Println("--- Running test: UploadDirectory ---")
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

	bucketName := fmt.Sprintf("%s-integration-test-bucket-upload-dir", projectID)
	destinationDir := "test-dir"

	// Create a bucket for the test
	err = csClient.CreateBucket(ctx, projectID, bucketName)
	if err != nil {
		log.Fatalf("Failed to create bucket: %v", err)
	}

	// Clean up the bucket
	defer func() {
		log.Println("Cleaning up bucket...")
		err = csClient.DeleteBucket(ctx, bucketName)
		if err != nil {
			log.Printf("Failed to delete bucket: %v", err)
		}
	}()

	tmpDir, err := os.MkdirTemp("", "test-dir-*")
	if err != nil {
		log.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		log.Fatalf("Failed to create subdirectory: %v", err)
	}

	tmpFile, err := os.Create(filepath.Join(subDir, "test-file.txt"))
	if err != nil {
		log.Fatalf("Failed to create temporary file: %v", err)
	}
	if _, err := tmpFile.Write([]byte("hello world")); err != nil {
		log.Fatalf("Failed to write to temporary file: %v", err)
	}
	tmpFile.Close()

	args := map[string]any{
		"bucket_name":     bucketName,
		"destination_dir": destinationDir,
		"source_path":     tmpDir,
	}

	var req mcp.CallToolRequest
	req.Params.Name = "cloudstorage.upload_directory"
	req.Params.Arguments = args

	log.Println("Calling tool 'cloudstorage.upload_directory'...")

	resp, err := mcpClient.CallTool(ctx, req)
	if err != nil {
		log.Fatalf("Tool call failed: %v", err)
	}

	if resp.IsError {
		log.Fatalf("Tool returned an error: %v", resp.Content)
	}

	log.Println("Tool call successful.")

	// Clean up the bucket and directory
	defer func() {
		log.Println("Cleaning up directory...")
		objectName := fmt.Sprintf("%s/subdir/test-file.txt", destinationDir)
		err := csClient.DeleteObject(ctx, bucketName, objectName)
		if err != nil {
			log.Printf("Failed to delete object: %v", err)
		}
	}()

	// Verify that the file was uploaded
	log.Println("Verifying directory upload...")
	objectName := fmt.Sprintf("%s/subdir/test-file.txt", destinationDir)
	err = csClient.CheckObjectExists(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("Failed to get object after upload: %v", err)
	}

	log.Println("Directory upload verification successful.")
}

func testCreateService(ctx context.Context, crClient cloudrunclient.CloudRunClient) {
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
		"image_url":     "us-docker.pkg.dev/cloudrun/container/hello",
		"port":          8080,
	}

	var req mcp.CallToolRequest
	req.Params.Name = "cloudrun.create_service"
	req.Params.Arguments = args

	log.Println("Calling tool 'cloudrun.create_service'...")

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

func testCreateRevision(ctx context.Context, crClient cloudrunclient.CloudRunClient) {
	log.Println("--- Running test: CreateRevision ---")
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

	// Create a service for the test
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
	req.Params.Name = "cloudrun.create_revision"
	req.Params.Arguments = args

	log.Println("Calling tool 'cloudrun.create_revision'...")

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

func testCreateServiceFromSource(ctx context.Context, crClient cloudrunclient.CloudRunClient) {
	log.Println("--- Running test: CreateServiceFromSource ---")
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
		"project_id":    projectID,
		"location":      "us-central1",
		"service_name":  "integration-test-service-from-source",
		"source":        tmpDir,
		"port":          8080,
	}

	var req mcp.CallToolRequest
	req.Params.Name = "cloudrun.create_service_from_source"
	req.Params.Arguments = args

	log.Println("Calling tool 'cloudrun.create_service_from_source'...")

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

