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

	_ "embed"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"devops-mcp-server/artifactregistry"
	"devops-mcp-server/cloudbuild"
	"devops-mcp-server/clouddeploy"
	"devops-mcp-server/cloudrun"
	"devops-mcp-server/cloudstorage"
	"devops-mcp-server/containeranalysis"
	"devops-mcp-server/devconnect"
	"devops-mcp-server/iam"
	"devops-mcp-server/prompts"
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
	if err := addArtifactRegistryTools(ctx, server); err != nil {
		return err
	}
	if err := addCloudBuildTools(ctx, server); err != nil {
		return err
	}
	if err := addCloudDeployTools(ctx, server); err != nil {
		return err
	}
	if err := addCloudRunTools(ctx, server); err != nil {
		return err
	}
	if err := addContainerAnalysisTools(ctx, server); err != nil {
		return err
	}
	if err := addDevConnectTools(ctx, server); err != nil {
		return err
	}
	if err := addIAMTools(ctx, server); err != nil {
		return err
	}
	if err := addCloudStorageTools(ctx, server); err != nil {
		return err
	}
	return nil
}

func addCloudStorageTools(ctx context.Context, server *mcp.Server) error {
	gcs, err := cloudstorage.NewClient(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to create Cloud Storage client: %v", err)
	}
	type createBucketArgs struct {
		ProjectID  string `json:"project_id"`
		BucketName string `json:"bucket_name"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudstorage.create_bucket", Description: "Creates a new Cloud Storage bucket."}, func(ctx context.Context, req *mcp.CallToolRequest, args createBucketArgs) (*mcp.CallToolResult, any, error) {
		err := gcs.CreateBucket(ctx, args.ProjectID, args.BucketName)
		return &mcp.CallToolResult{}, nil, err
	})
	type uploadFileArgs struct {
		ProjectID  string `json:"project_id"`
		BucketName string `json:"bucket_name"`
		ObjectName string `json:"object_name"`
		FilePath   string `json:"file_path"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudstorage.upload_file", Description: "Uploads a file to a Cloud Storage bucket."}, func(ctx context.Context, req *mcp.CallToolRequest, args uploadFileArgs) (*mcp.CallToolResult, any, error) {
		err := gcs.UploadFile(ctx, args.ProjectID, args.BucketName, args.ObjectName, args.FilePath)
		return &mcp.CallToolResult{}, nil, err
	})
	type uploadDirectoryArgs struct {
		ProjectID      string `json:"project_id"`
		BucketName     string `json:"bucket_name"`
		DestinationDir string `json:"destination_dir"`
		SourcePath     string `json:"source_path"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudstorage.upload_directory", Description: "Uploads a directory to a Cloud Storage bucket."}, func(ctx context.Context, req *mcp.CallToolRequest, args uploadDirectoryArgs) (*mcp.CallToolResult, any, error) {
		err := gcs.UploadDirectory(ctx, args.ProjectID, args.BucketName, args.DestinationDir, args.SourcePath)
		return &mcp.CallToolResult{}, nil, err
	})
	return nil
}

func addArtifactRegistryTools(ctx context.Context, server *mcp.Server) error {
	c, err := artifactregistry.NewGRPCClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create artifact registry client: %v", err)
	}
	ar, err := artifactregistry.NewClient(c)
	if err != nil {
		return fmt.Errorf("failed to create Artifact Registry client: %v", err)
	}
	type createRepoArgs struct {
		ProjectID    string `json:"project_id"`
		Location     string `json:"location"`
		RepositoryID string `json:"repository_id"`
		Format       string `json:"format"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "artifactregistry.create_repository", Description: "Creates a new Artifact Registry repository."}, func(ctx context.Context, req *mcp.CallToolRequest, args createRepoArgs) (*mcp.CallToolResult, any, error) {
		res, err := ar.CreateRepository(ctx, args.ProjectID, args.Location, args.RepositoryID, args.Format)
		return &mcp.CallToolResult{}, res, err
	})
	// type getRepoArgs struct {
	// 	ProjectID    string `json:"project_id"`
	// 	Location     string `json:"location"`
	// 	RepositoryID string `json:"repository_id"`
	// }
	// mcp.AddTool(server, &mcp.Tool{Name: "artifactregistry.get_repository", Description: "Gets an Artifact Registry repository."}, func(ctx context.Context, req *mcp.CallToolRequest, args getRepoArgs) (*mcp.CallToolResult, any, error) {
	// 	res, err := ar.GetRepository(ctx, args.ProjectID, args.Location, args.RepositoryID)
	// 	return &mcp.CallToolResult{}, res, err
	// })
	return nil
}

func addCloudBuildTools(ctx context.Context, server *mcp.Server) error {
	cb, err := cloudbuild.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Cloud Build client: %v", err)
	}
	type createTriggerArgs struct {
		ProjectID      string `json:"project_id"`
		Location       string `json:"location"`
		TriggerID      string `json:"trigger_id"`
		RepoLink       string `json:"repo_link"`
		ServiceAccount string `json:"service_account"`
		Branch         string `json:"branch"`
		Tag            string `json:"tag"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudbuild.create_trigger", Description: "Creates a new Cloud Build trigger."}, func(ctx context.Context, req *mcp.CallToolRequest, args createTriggerArgs) (*mcp.CallToolResult, any, error) {
		res, err := cb.CreateTrigger(ctx, args.ProjectID, args.Location, args.TriggerID, args.RepoLink, args.ServiceAccount, args.Branch, args.Tag)
		return &mcp.CallToolResult{}, res, err
	})
	type runTriggerArgs struct {
		ProjectID string `json:"project_id"`
		Location  string `json:"location"`
		TriggerID string `json:"trigger_id"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudbuild.run_trigger", Description: "Runs a Cloud Build trigger."}, func(ctx context.Context, req *mcp.CallToolRequest, args runTriggerArgs) (*mcp.CallToolResult, any, error) {
		res, err := cb.RunTrigger(ctx, args.ProjectID, args.Location, args.TriggerID)
		return &mcp.CallToolResult{}, res, err
	})
	type listTriggersArgs struct {
		ProjectID string `json:"project_id"`
		Location  string `json:"location"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudbuild.list_triggers", Description: "Lists all Cloud Build triggers in a given location."}, func(ctx context.Context, req *mcp.CallToolRequest, args listTriggersArgs) (*mcp.CallToolResult, any, error) {
		res, err := cb.ListTriggers(ctx, args.ProjectID, args.Location)
		return &mcp.CallToolResult{}, res, err
	})
	type buildContainerArgs struct {
		ProjectID      string `json:"project_id"`
		Location       string `json:"location"`
		Repository     string `json:"repository"`
		ImageName      string `json:"image_name"`
		Tag            string `json:"tag"`
		DockerfilePath string `json:"dockerfile_path"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudbuild.build_container", Description: "Builds a container image using Cloud Build."}, func(ctx context.Context, req *mcp.CallToolRequest, args buildContainerArgs) (*mcp.CallToolResult, any, error) {
		res, err := cb.BuildContainer(ctx, args.ProjectID, args.Location, args.Repository, args.ImageName, args.Tag, args.DockerfilePath)
		return &mcp.CallToolResult{}, res, err
	})
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

func addCloudRunTools(ctx context.Context, server *mcp.Server) error {
	cr, err := cloudrun.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Cloud Run client: %v", err)
	}
	type createServiceArgs struct {
		ProjectID   string `json:"project_id"`
		Location    string `json:"location"`
		ServiceName string `json:"service_name"`
		ImageURL    string `json:"image_url"`
		Port        int32  `json:"port"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudrun.create_service", Description: "Creates a new Cloud Run service."}, func(ctx context.Context, req *mcp.CallToolRequest, args createServiceArgs) (*mcp.CallToolResult, any, error) {
		res, err := cr.CreateService(ctx, args.ProjectID, args.Location, args.ServiceName, args.ImageURL, args.Port)
		return &mcp.CallToolResult{}, res, err
	})
	type createRevisionArgs struct {
		ProjectID    string `json:"project_id"`
		Location     string `json:"location"`
		ServiceName  string `json:"service_name"`
		ImageURL     string `json:"image_url"`
		RevisionName string `json:"revision_name"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudrun.create_revision", Description: "Creates a new Cloud Run revision for a service with a new Docker image."}, func(ctx context.Context, req *mcp.CallToolRequest, args createRevisionArgs) (*mcp.CallToolResult, any, error) {
		res, err := cr.CreateRevision(ctx, args.ProjectID, args.Location, args.ServiceName, args.ImageURL, args.RevisionName)
		return &mcp.CallToolResult{}, res, err
	})
	type deployFromSourceArgs struct {
		ProjectID   string `json:"project_id"`
		Location    string `json:"location"`
		ServiceName string `json:"service_name"`
		Source      string `json:"source"`
		Port        int32  `json:"port"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudrun.deploy_from_source", Description: "Deploys a new Cloud Run service or updates an existing one from source."}, func(ctx context.Context, req *mcp.CallToolRequest, args deployFromSourceArgs) (*mcp.CallToolResult, any, error) {
		err := cr.DeployFromSource(ctx, args.ProjectID, args.Location, args.ServiceName, args.Source, args.Port)
		return &mcp.CallToolResult{}, nil, err
	})
	type deployFromImageArgs struct {
		ProjectID   string `json:"project_id"`
		Location    string `json:"location"`
		ServiceName string `json:"service_name"`
		ImageURL    string `json:"image_url"`
		Port        int32  `json:"port"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudrun.deploy_from_image", Description: "Deploys a new Cloud Run service from a container image."}, func(ctx context.Context, req *mcp.CallToolRequest, args deployFromImageArgs) (*mcp.CallToolResult, any, error) {
		err := cr.DeployFromImage(ctx, args.ProjectID, args.Location, args.ServiceName, args.ImageURL, args.Port)
		return &mcp.CallToolResult{}, nil, err
	})
	return nil
}

func addContainerAnalysisTools(ctx context.Context, server *mcp.Server) error {
	ca, err := containeranalysis.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Container Analysis client: %v", err)
	}
	type listVulnerabilitiesArgs struct {
		ProjectID   string `json:"project_id"`
		ResourceURL string `json:"resource_url"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "containeranalysis.list_vulnerabilities", Description: "Lists vulnerabilities for a given image resource URL using Container Analysis."}, func(ctx context.Context, req *mcp.CallToolRequest, args listVulnerabilitiesArgs) (*mcp.CallToolResult, any, error) {
		res, err := ca.ListVulnerabilities(ctx, args.ProjectID, args.ResourceURL)
		return &mcp.CallToolResult{}, res, err
	})
	return nil
}

func addDevConnectTools(ctx context.Context, server *mcp.Server) error {
	dc, err := devconnect.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Dev Connect client: %v", err)
	}
	type createConnectionArgs struct {
		ProjectID    string `json:"project_id"`
		Location     string `json:"location"`
		ConnectionID string `json:"connection_id"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "devconnect.create_connection", Description: "Creates a new Developer Connect connection."}, func(ctx context.Context, req *mcp.CallToolRequest, args createConnectionArgs) (*mcp.CallToolResult, any, error) {
		res, err := dc.CreateConnection(ctx, args.ProjectID, args.Location, args.ConnectionID)
		return &mcp.CallToolResult{}, res, err
	})
	type createGitRepositoryLinkArgs struct {
		ProjectID    string `json:"project_id"`
		Location     string `json:"location"`
		ConnectionID string `json:"connection_id"`
		RepoLinkID   string `json:"repo_link_id"`
		RepoURI      string `json:"repo_uri"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "devconnect.create_git_repository_link", Description: "Creates a new Developer Connect Git Repository Link."}, func(ctx context.Context, req *mcp.CallToolRequest, args createGitRepositoryLinkArgs) (*mcp.CallToolResult, any, error) {
		res, err := dc.CreateGitRepositoryLink(ctx, args.ProjectID, args.Location, args.ConnectionID, args.RepoLinkID, args.RepoURI)
		return &mcp.CallToolResult{}, res, err
	})
	type listConnectionsArgs struct {
		ProjectID string `json:"project_id"`
		Location  string `json:"location"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "devconnect.list_connections", Description: "Lists Developer Connect connections."}, func(ctx context.Context, req *mcp.CallToolRequest, args listConnectionsArgs) (*mcp.CallToolResult, any, error) {
		res, err := dc.ListConnections(ctx, args.ProjectID, args.Location)
		return &mcp.CallToolResult{}, res, err
	})
	type getConnectionArgs struct {
		ProjectID    string `json:"project_id"`
		Location     string `json:"location"`
		ConnectionID string `json:"connection_id"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "devconnect.get_connection", Description: "Gets a Developer Connect connection."}, func(ctx context.Context, req *mcp.CallToolRequest, args getConnectionArgs) (*mcp.CallToolResult, any, error) {
		res, err := dc.GetConnection(ctx, args.ProjectID, args.Location, args.ConnectionID)
		return &mcp.CallToolResult{}, res, err
	})
	type findGitRepositoryLinksForGitRepoArgs struct {
		ProjectID string `json:"project_id"`
		Location  string `json:"location"`
		RepoURI   string `json:"repo_uri"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "devconnect.find_git_repository_links_for_git_repo", Description: "Finds already configured Developer Connect Git Repository Links for a particular git repository."}, func(ctx context.Context, req *mcp.CallToolRequest, args findGitRepositoryLinksForGitRepoArgs) (*mcp.CallToolResult, any, error) {
		res, err := dc.FindGitRepositoryLinksForGitRepo(ctx, args.ProjectID, args.Location, args.RepoURI)
		return &mcp.CallToolResult{}, res, err
	})
	return nil
}

func addIAMTools(ctx context.Context, server *mcp.Server) error {
	iamClient, err := iam.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create IAM client: %v", err)
	}
	type createServiceAccountArgs struct {
		ProjectID   string `json:"project_id"`
		DisplayName string `json:"display_name"`
		AccountID   string `json:"account_id"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "iam.create_service_account", Description: "Creates a new Google Cloud Platform service account."}, func(ctx context.Context, req *mcp.CallToolRequest, args createServiceAccountArgs) (*mcp.CallToolResult, any, error) {
		res, err := iamClient.CreateServiceAccount(ctx, args.ProjectID, args.DisplayName, args.AccountID)
		return &mcp.CallToolResult{}, res, err
	})
	type addIAMRoleBindingArgs struct {
		ResourceID string `json:"resource_id"`
		Role       string `json:"role"`
		Member     string `json:"member"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "iam.add_iam_role_binding", Description: "Adds an IAM role binding to a Google Cloud Platform resource."}, func(ctx context.Context, req *mcp.CallToolRequest, args addIAMRoleBindingArgs) (*mcp.CallToolResult, any, error) {
		res, err := iamClient.AddIAMRoleBinding(ctx, args.ResourceID, args.Role, args.Member)
		return &mcp.CallToolResult{}, res, err
	})
	type listServiceAccountsArgs struct {
		ProjectID string `json:"project_id"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "iam.list_service_accounts", Description: "Lists all service accounts in a project."}, func(ctx context.Context, req *mcp.CallToolRequest, args listServiceAccountsArgs) (*mcp.CallToolResult, any, error) {
		res, err := iamClient.ListServiceAccounts(ctx, args.ProjectID)
		return &mcp.CallToolResult{}, res, err
	})
	type getIAMRoleBindingArgs struct {
		ProjectID           string `json:"project_id"`
		ServiceAccountEmail string `json:"service_account_email"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "iam.get_iam_role_binding", Description: "Gets the IAM role bindings for a service account."}, func(ctx context.Context, req *mcp.CallToolRequest, args getIAMRoleBindingArgs) (*mcp.CallToolResult, any, error) {
		res, err := iamClient.GetIAMRoleBinding(ctx, args.ProjectID, args.ServiceAccountEmail)
		return &mcp.CallToolResult{}, res, err
	})
	return nil
}
