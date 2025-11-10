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
	"devops-mcp-server/cloudbuild"
	"devops-mcp-server/cloudrun"
	"devops-mcp-server/cloudstorage"
	"devops-mcp-server/devconnect"
	"devops-mcp-server/osv"
	"devops-mcp-server/prompts"

	artifactregistryclient "devops-mcp-server/artifactregistry/client"
	cloudbuildclient "devops-mcp-server/cloudbuild/client"
	cloudrunclient "devops-mcp-server/cloudrun/client"
	cloudstorageclient "devops-mcp-server/cloudstorage/client"
	developerconnectclient "devops-mcp-server/devconnect/client"
	iamclient "devops-mcp-server/iam/client"
	osvclient "devops-mcp-server/osv/client"
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

	addAllPrompts(ctx, server)

	return server
}

func addAllPrompts(ctx context.Context, server *mcp.Server) {
	// Add design prompt.
	prompts.DesignPrompt(ctx, server)
	// Add deploy prompt.
	prompts.DeployPrompt(ctx, server)
}

func addAllTools(ctx context.Context, server *mcp.Server) error {
	i, err := iamclient.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create IAM client: %w", err)
	}

	ctxWithDeps := iamclient.ContextWithClient(ctx, i)

	r, err := resourcemanagerclient.NewClient(ctxWithDeps)
	if err != nil {
		return fmt.Errorf("failed to create resource manager client: %w", err)
	}

	ctxWithDeps = resourcemanagerclient.ContextWithClient(ctxWithDeps, r)

	arClient, err := artifactregistryclient.NewArtifactRegistryClient(ctxWithDeps)
	if err != nil {
		return fmt.Errorf("failed to create ArtifactRegistry client: %w", err)
	}
	ctxWithDeps = artifactregistryclient.ContextWithClient(ctxWithDeps, arClient)

	if err := artifactregistry.AddTools(ctxWithDeps, server); err != nil {
		return err
	}

	crClient, err := cloudrunclient.NewCloudRunClient(ctxWithDeps)
	if err != nil {
		return fmt.Errorf("failed to create CloudRun client: %w", err)
	}
	ctxWithDeps = cloudrunclient.ContextWithClient(ctxWithDeps, crClient)

	if err := cloudrun.AddTools(ctxWithDeps, server); err != nil {
		return err
	}
	devConnectClient, err := developerconnectclient.NewDeveloperConnectClient(ctxWithDeps)
	if err != nil {
		return fmt.Errorf("failed to create dev connect client: %w", err)
	}
	ctxWithDeps = developerconnectclient.ContextWithClient(ctxWithDeps, devConnectClient)

	if err := devconnect.AddTools(ctxWithDeps, server); err != nil {
		return err
	}

	csClient, err := cloudstorageclient.NewCloudStorageClient(ctxWithDeps)
	if err != nil {
		return fmt.Errorf("failed to create CloudStorage client: %w", err)
	}
	ctxWithDeps = cloudstorageclient.ContextWithClient(ctxWithDeps, csClient)

	if err := cloudstorage.AddTools(ctxWithDeps, server); err != nil {
		return err
	}
	cbClient, err := cloudbuildclient.NewCloudBuildClient(ctxWithDeps)
	if err != nil {
		return fmt.Errorf("failed to create CloudBuild client: %w", err)
	}
	ctxWithDeps = cloudbuildclient.ContextWithClient(ctxWithDeps, cbClient)

	if err := cloudbuild.AddTools(ctxWithDeps, server); err != nil {
		return err
	}

	osvClient, err := osvclient.NewClient(ctxWithDeps)
	if err != nil {
		return fmt.Errorf("failed to create OSV client: %w", err)
	}
	ctxWithDeps = osvclient.ContextWithClient(ctxWithDeps, osvClient)

	if err := osv.AddTools(ctxWithDeps, server); err != nil {
		return err
	}
	return nil
}
