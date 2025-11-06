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
	"devops-mcp-server/clouddeploy"
	"devops-mcp-server/cloudrun"
	"devops-mcp-server/cloudstorage"
	"devops-mcp-server/devconnect"
	"devops-mcp-server/prompts"

	artifactregistryclient "devops-mcp-server/artifactregistry/client"
	cloudbuildclient "devops-mcp-server/cloudbuild/client"
	cloudrunclient "devops-mcp-server/cloudrun/client"
	cloudstorageclient "devops-mcp-server/cloudstorage/client"
	developerconnectclient "devops-mcp-server/devconnect/client"
	iamclient "devops-mcp-server/iam/client"
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
	if err := addCloudDeployTools(ctx, server); err != nil {
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
	return nil
}

func addCloudDeployTools(ctx context.Context, server *mcp.Server) error {
	cd, err := clouddeploy.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Cloud Deploy client: %v", err)
	}
	type createDeliveryPipelineArgs struct {
		ProjectID   string `json:"project_id"`
		Location    string `json:"location"`
		PipelineID  string `json:"pipeline_id"`
		Description string `json:"description"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "clouddeploy.create_delivery_pipeline", Description: "Creates a new Cloud Deploy delivery pipeline."}, func(ctx context.Context, req *mcp.CallToolRequest, args createDeliveryPipelineArgs) (*mcp.CallToolResult, any, error) {
		res, err := cd.CreateDeliveryPipeline(ctx, args.ProjectID, args.Location, args.PipelineID, args.Description)
		return &mcp.CallToolResult{}, res, err
	})
	type createGKETargetArgs struct {
		ProjectID   string `json:"project_id"`
		Location    string `json:"location"`
		TargetID    string `json:"target_id"`
		GKECluster  string `json:"gke_cluster"`
		Description string `json:"description"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "clouddeploy.create_gke_target", Description: "Creates a new Cloud Deploy GKE target."}, func(ctx context.Context, req *mcp.CallToolRequest, args createGKETargetArgs) (*mcp.CallToolResult, any, error) {
		res, err := cd.CreateGKETarget(ctx, args.ProjectID, args.Location, args.TargetID, args.GKECluster, args.Description)
		return &mcp.CallToolResult{}, res, err
	})
	type createCloudRunTargetArgs struct {
		ProjectID   string `json:"project_id"`
		Location    string `json:"location"`
		TargetID    string `json:"target_id"`
		Description string `json:"description"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "clouddeploy.create_cloud_run_target", Description: "Creates a new Cloud Deploy Cloud Run target."}, func(ctx context.Context, req *mcp.CallToolRequest, args createCloudRunTargetArgs) (*mcp.CallToolResult, any, error) {
		res, err := cd.CreateCloudRunTarget(ctx, args.ProjectID, args.Location, args.TargetID, args.Description)
		return &mcp.CallToolResult{}, res, err
	})
	type createRolloutArgs struct {
		ProjectID  string `json:"project_id"`
		Location   string `json:"location"`
		PipelineID string `json:"pipeline_id"`
		ReleaseID  string `json:"release_id"`
		RolloutID  string `json:"rollout_id"`
		TargetID   string `json:"target_id"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "clouddeploy.create_rollout", Description: "Creates a new Cloud Deploy rollout."}, func(ctx context.Context, req *mcp.CallToolRequest, args createRolloutArgs) (*mcp.CallToolResult, any, error) {
		res, err := cd.CreateRollout(ctx, args.ProjectID, args.Location, args.PipelineID, args.ReleaseID, args.RolloutID, args.TargetID)
		return &mcp.CallToolResult{}, res, err
	})
	type listDeliveryPipelinesArgs struct {
		ProjectID string `json:"project_id"`
		Location  string `json:"location"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "clouddeploy.list_delivery_pipelines", Description: "Lists all Cloud Deploy delivery pipelines."}, func(ctx context.Context, req *mcp.CallToolRequest, args listDeliveryPipelinesArgs) (*mcp.CallToolResult, any, error) {
		res, err := cd.ListDeliveryPipelines(ctx, args.ProjectID, args.Location)
		return &mcp.CallToolResult{}, res, err
	})
	type listTargetsArgs struct {
		ProjectID string `json:"project_id"`
		Location  string `json:"location"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "clouddeploy.list_targets", Description: "Lists all Cloud Deploy targets."}, func(ctx context.Context, req *mcp.CallToolRequest, args listTargetsArgs) (*mcp.CallToolResult, any, error) {
		res, err := cd.ListTargets(ctx, args.ProjectID, args.Location)
		return &mcp.CallToolResult{}, res, err
	})
	type listReleasesArgs struct {
		ProjectID  string `json:"project_id"`
		Location   string `json:"location"`
		PipelineID string `json:"pipeline_id"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "clouddeploy.list_releases", Description: "Lists all Cloud Deploy releases for a given delivery pipeline."}, func(ctx context.Context, req *mcp.CallToolRequest, args listReleasesArgs) (*mcp.CallToolResult, any, error) {
		res, err := cd.ListReleases(ctx, args.ProjectID, args.Location, args.PipelineID)
		return &mcp.CallToolResult{}, res, err
	})
	type listRolloutsArgs struct {
		ProjectID  string `json:"project_id"`
		Location   string `json:"location"`
		PipelineID string `json:"pipeline_id"`
		ReleaseID  string `json:"release_id"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "clouddeploy.list_rollouts", Description: "Lists all Cloud Deploy rollouts for a given release."}, func(ctx context.Context, req *mcp.CallToolRequest, args listRolloutsArgs) (*mcp.CallToolResult, any, error) {
		res, err := cd.ListRollouts(ctx, args.ProjectID, args.Location, args.PipelineID, args.ReleaseID)
		return &mcp.CallToolResult{}, res, err
	})
	return nil
}
