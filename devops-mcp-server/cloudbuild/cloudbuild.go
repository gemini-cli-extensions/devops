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

package cloudbuild

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	cloudbuildclient "devops-mcp-server/cloudbuild/client"
	iamclient "devops-mcp-server/iam/client"
	resourcemanagerclient "devops-mcp-server/resourcemanager/client"
)

// Handler holds the clients for the cloudbuild service.
type Handler struct {
	CbClient cloudbuildclient.CloudBuildClient
	IClient  iamclient.IAMClient
	RClient  resourcemanagerclient.ResourcemanagerClient
}

// Register registers the cloudbuild tools with the MCP server.
func (h *Handler) Register(server *mcp.Server) {
	addCreateTriggerTool(server, h.CbClient, h.IClient, h.RClient)
	addRunTriggerTool(server, h.CbClient)
	addListTriggersTool(server, h.CbClient)
}

type RunTriggerArgs struct {
	ProjectID string `json:"project_id" jsonschema:"The Google Cloud project ID."`
	Location  string `json:"location" jsonschema:"The Google Cloud location for the trigger."`
	TriggerID string `json:"trigger_id" jsonschema:"The ID of the trigger."`
	Branch    string `json:"branch,omitempty" jsonschema:"The branch to to run the trigger at. Should be regex e.g. '^main$'"`
	Tag       string `json:"tag,omitempty" jsonschema:"The tag to to run the trigger at. Should be regex e.g. '^nightly$'"`
	CommitSha string `json:"commit_sha,omitempty" jsonschema:"The commit sha to run the trigger at. Exact commit sha e.g. 12ede13"`
}

var runTriggerToolFunc func(ctx context.Context, req *mcp.CallToolRequest, args RunTriggerArgs) (*mcp.CallToolResult, any, error)

func addRunTriggerTool(server *mcp.Server, cbClient cloudbuildclient.CloudBuildClient) {
	runTriggerToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args RunTriggerArgs) (*mcp.CallToolResult, any, error) {
		res, err := cbClient.RunBuildTrigger(ctx, args.ProjectID, args.Location, args.TriggerID, args.Branch, args.Tag, args.CommitSha)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to run trigger: %w", err)
		}
		return &mcp.CallToolResult{}, res, nil
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudbuild.run_trigger", Description: "Runs a Cloud Build trigger."}, runTriggerToolFunc)
}

type ListTriggersArgs struct {
	ProjectID string `json:"project_id" jsonschema:"The Google Cloud project ID."`
	Location  string `json:"location" jsonschema:"The Google Cloud location for the triggers."`
}

var listTriggersToolFunc func(ctx context.Context, req *mcp.CallToolRequest, args ListTriggersArgs) (*mcp.CallToolResult, any, error)

func addListTriggersTool(server *mcp.Server, cbClient cloudbuildclient.CloudBuildClient) {
	listTriggersToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args ListTriggersArgs) (*mcp.CallToolResult, any, error) {
		res, err := cbClient.ListBuildTriggers(ctx, args.ProjectID, args.Location)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to list triggers: %w", err)
		}
		return &mcp.CallToolResult{}, map[string]any{"triggers": res}, nil
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudbuild.list_triggers", Description: "Lists all Cloud Build triggers in a given location."}, listTriggersToolFunc)
}

type CreateTriggerArgs struct {
	ProjectID      string `json:"project_id" jsonschema:"The Google Cloud project ID."`
	Location       string `json:"location" jsonschema:"The Google Cloud location for the trigger."`
	TriggerID      string `json:"trigger_id" jsonschema:"The ID of the trigger."`
	RepoLink       string `json:"repo_link" jsonschema:"The Developer Connect repository link, use dev connect setup repo to create a connect and repo link"`
	ServiceAccount string `json:"service_account" jsonschema:"The service account to use for the build. E.g. serviceAccount:123-compute@developer.gserviceaccount.com"`
	Branch         string `json:"branch,omitempty" jsonschema:"Create builds on push to branch. Should be regex e.g. '^main$'"`
	Tag            string `json:"tag,omitempty" jsonschema:"Create builds on new tag push. Should be regex e.g. '^nightly$'"`
}

var createTriggerToolFunc func(ctx context.Context, req *mcp.CallToolRequest, args CreateTriggerArgs) (*mcp.CallToolResult, any, error)

func addCreateTriggerTool(server *mcp.Server, cbClient cloudbuildclient.CloudBuildClient, iamClient iamclient.IAMClient, rmClient resourcemanagerclient.ResourcemanagerClient) {
	createTriggerToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args CreateTriggerArgs) (*mcp.CallToolResult, any, error) {
		resolvedSA, err := setPermissionsForCloudBuildSA(ctx, args.ProjectID, args.ServiceAccount, rmClient, iamClient)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to grant necessary permissions for the Cloud build service account: %w", err)
		}
		res, err := cbClient.CreateBuildTrigger(ctx, args.ProjectID, args.Location, args.TriggerID, args.RepoLink, args.Branch, args.Tag, resolvedSA)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to create trigger: %w", err)
		}
		return &mcp.CallToolResult{}, res, nil
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudbuild.create_trigger", Description: "Creates a new Cloud Build trigger."}, createTriggerToolFunc)
}

// setPermissionsForSA resolves the SA (if default) and grants it a role.
// It creates and manages its own Resource Manager client.
func setPermissionsForCloudBuildSA(ctx context.Context, projectID, serviceAccount string, rmClient resourcemanagerclient.ResourcemanagerClient, iamClient iamclient.IAMClient) (string, error) {
	// Construct the Compute Engine default service account email
	resolvedSA := serviceAccount
	if resolvedSA == "" {
		projectNumber, err := rmClient.ToProjectNumber(ctx, projectID)
		if err != nil {
			return "", fmt.Errorf("unable to resolve project id to number: %w", err)
		}
		resolvedSA = fmt.Sprintf("serviceAccount:%d-compute@developer.gserviceaccount.com", projectNumber)
	}

	roles := []string{"roles/developerconnect.tokenAccessor"}
	for _, r := range roles {
		_, err := iamClient.AddIAMRoleBinding(ctx, fmt.Sprintf("projects/%s", projectID), r, resolvedSA)
		if err != nil {
			return "", fmt.Errorf("unable to add role %s to member %s on resource %s err: %w", r, resolvedSA, fmt.Sprintf("projects/%s", projectID), err)
		}
	}
	return resolvedSA, nil
}
