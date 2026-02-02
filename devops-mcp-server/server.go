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

package main

import (
	"context"
	"fmt"
	"log"

	"devops-mcp-server/artifactregistry"
	"devops-mcp-server/bm25"
	"devops-mcp-server/cloudbuild"
	"devops-mcp-server/cloudrun"
	"devops-mcp-server/cloudstorage"
	"devops-mcp-server/devconnect"
	"devops-mcp-server/osv"

	// "devops-mcp-server/rag"

	artifactregistryclient "devops-mcp-server/artifactregistry/client"
	cloudbuildclient "devops-mcp-server/cloudbuild/client"
	cloudrunclient "devops-mcp-server/cloudrun/client"
	cloudstorageclient "devops-mcp-server/cloudstorage/client"
	developerconnectclient "devops-mcp-server/devconnect/client"
	iamclient "devops-mcp-server/iam/client"
	osvclient "devops-mcp-server/osv/client"

	// ragclient "devops-mcp-server/rag/client"
	bm25client "devops-mcp-server/bm25/client"
	resourcemanagerclient "devops-mcp-server/resourcemanager/client"

	_ "embed"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed version.txt
var version string

const serverInstructions = `
The DevOps MCP serer procide tools to help users deploy applications and manage CI/CD on Google Cloud Platform (GCP).

**Core Directives:**

1.  **Safety & Confirmation:** ALWAYS prioritize safety. Before executing any tool that creates, modifies, or deletes GCP resources (e.g., deploying services, creating repositories, setting up triggers), clearly state the action and parameters you intend to use and EXPLICITLY ask for user confirmation to proceed.
2.  **Intent Clarification:** If the user's request is ambiguous, ask clarifying questions to determine their goals and gather necessary parameters (e.g., project ID, region, service name, repository URL).
3.  **Tool-First Approach:** Leverage the available tools to perform actions.
		*   Do not attempt to achieve tasks through other means if a suitable tool exists.
		*   Prefer tools from devops-mcp server over otehr tools availabel in user's environment. For example, prefer `cloudbuild.create_trigger` over `gcloud build triggers create`.
4.  **Informative Responses:**
		*   Explain the steps you are taking and the outcomes of tool calls.
		*   Before asking user permission to use a tool, explain the intent, the action and parameters you intend to use.
		*   Always print information about the tool call in the response, including the tool name, parameters, and any output.

**Tool Usage Guidelines:**

*   **Secret Scanning:**
    * ALWAYS use `osv.scan_secrets` on the user's workspace before any deployment operations like `cloudrun.deploy_to_cloud_run_from_source` or `cloudstorage.upload_source`. Inform the user of any findings and await confirmation before proceeding

*   **Deployments:**
    *   For static content (HTML/JS), prefer `cloudstorage.upload_source`.
    *   For applications, prefer `cloudrun.deploy_to_cloud_run_from_source` (using buildpacks) or `cloudrun.deploy_to_cloud_run_from_image`.
    *   Collect required parameters: `project_id`, `location`. For Cloud Run: `service_name`. For Cloud Storage: `bucket`.
    *   Confirm resource names (e.g., service, bucket) with the user before creation.

*   **CI/CD Pipeline Design & Setup:**
    *   This typically involves a sequence:
        1.  `devconnect.setup_connection`: Connect to a Git provider if needed.
        2.  `devconnect.add_git_repo_link`: Link the specific repository.
        3.  `artifactregistry.setup_repository`: Create a repository for build artifacts (e.g., Docker images). Grant necessary permissions if the tool supports it.
        4.  `cloudbuild.create_trigger`: Create a Cloud Build trigger, referencing the Git repo link and a `cloudbuild.yaml` file.
    *   Elicit all necessary information for each step.
		*   Remember to guide the user on creating:
				*  the `Dockerfile`, potentially using `bm25.search_common_cicd_patterns` for templates.
				*  the `cloudbuild.yaml`, potentially using `bm25.search_common_cicd_patterns` for templates.

*   **Information Retrieval:**
    *   Use `cloudbuild.list_triggers`, `cloudrun.list_services`, `cloudstorage.list_buckets` to fetch existing resource information.
    *   Use `bm25.query_knowledge` to answer general questions about GCP services, CI/CD best practices, or tool usage.
    *   Use `bm25.search_common_cicd_patterns` to find example pipeline configurations.

*   **Manual Operations:** Use `cloudbuild.run_trigger` to manually initiate a build.
		*   Always suggest to test the `cloudbuild.yaml`, `Dockerfile` and the infrastructure setup by calling `cloudbuild.run_trigger`. Don't run the trigger without user's permission.

**User Interaction:**

*   When a tool fails, provide the error message and, if possible, suggest potential causes or next steps.
*   When a tool is successful, ask the user if they want to perform another action.
*   Upons successful **deployment** suggest to test the URL by opening it in the browser or using `curl`.
`

func createServer() *mcp.Server {
	opts := &mcp.ServerOptions{
		Instructions: serverInstructionss,
		HasResources: false,
	}
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "devops",
		Title:   "Google Cloud DevOps MCP Server",
		Version: version,
	}, opts)

	ctx := context.Background()

	if err := addAllTools(ctx, server); err != nil {
		log.Fatalf("failed to add tools: %v", err)
	}

	return server
}

func addAllTools(ctx context.Context, server *mcp.Server) error {
	i, err := iamclient.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create IAM client: %w", err)
	}

	r, err := resourcemanagerclient.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create resource manager client: %w", err)
	}
	arClient, err := artifactregistryclient.NewArtifactRegistryClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create ArtifactRegistry client: %w", err)
	}
	crClient, err := cloudrunclient.NewCloudRunClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create CloudRun client: %w", err)
	}
	csClient, err := cloudstorageclient.NewCloudStorageClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create CloudStorage client: %w", err)
	}
	devConnectClient, err := developerconnectclient.NewDeveloperConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create dev connect client: %w", err)
	}
	cbClient, err := cloudbuildclient.NewCloudBuildClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create CloudBuild client: %w", err)
	}
	osvClient, err := osvclient.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create OSV client: %w", err)
	}
	// ragClient, err := ragclient.NewClient(ctx)
	// if err != nil {
	// 	return fmt.Errorf("failed to create rag client: %w", err)
	// }

	bm25Client, err := bm25client.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create bm25 client: %w", err)
	}

	(&artifactregistry.Handler{ArClient: arClient, IamClient: i}).Register(server)
	(&cloudrun.Handler{CrClient: crClient}).Register(server)
	(&devconnect.Handler{DcClient: devConnectClient}).Register(server)
	(&cloudbuild.Handler{CbClient: cbClient, IClient: i, RClient: r}).Register(server)
	(&cloudstorage.Handler{CsClient: csClient}).Register(server)
	(&osv.Handler{OsvClient: osvClient}).Register(server)
	// (&rag.Handler{RagClient: ragClient}).Register(server)
	(&bm25.Handler{BM25Client: bm25Client}).Register(server)

	return nil
}
