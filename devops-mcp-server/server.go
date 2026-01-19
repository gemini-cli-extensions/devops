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

func createServer() *mcp.Server {
	opts := &mcp.ServerOptions{
		Instructions: "Google Cloud DevOps MCP Server",
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
