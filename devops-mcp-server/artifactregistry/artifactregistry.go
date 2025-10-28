// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package artifactregistry

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	artifactregistryclient "devops-mcp-server/artifactregistry/client"
	iamclient "devops-mcp-server/iam/client"
)

// AddTools adds all artifact registry related tools to the mcp server.
// It expects the artifactregistryclient and mcp.Server to be in the context.
func AddTools(ctx context.Context, server *mcp.Server) error {
	a, ok := artifactregistryclient.ClientFrom(ctx)
	if !ok {
		return fmt.Errorf("artifact registry client not found in context")
	}
	i, ok := iamclient.ClientFrom(ctx)
	if !ok {
		return fmt.Errorf("artifact registry client not found in context")
	}
	addSetupRepositoryTool(server, a, i)
	return nil
}

type SetupRepoArgs struct {
	ProjectID           string `json:"project_id" jsonschema:"The Google Cloud project ID."`
	Location            string `json:"location" jsonschema:"The Google Cloud location for the repository."`
	RepositoryID        string `json:"repository_id" jsonschema:"The ID of the repository."`
	Format              string `json:"format" jsonschema:"The format of the repository (e.g., DOCKER)."`
	ServiceAccountEmail string `json:"service_account_email,omitempty" jsonschema:"The email of the service account to grant Artifact Registry Writer permissions to."`
}

var setupRepositoryToolFunc func(ctx context.Context, req *mcp.CallToolRequest, args SetupRepoArgs) (*mcp.CallToolResult, any, error)

func addSetupRepositoryTool(server *mcp.Server, arClient artifactregistryclient.ArtifactRegistryClient, iamClient iamclient.IAMClient) {
	setupRepositoryToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args SetupRepoArgs) (*mcp.CallToolResult, any, error) {
		res, err := arClient.CreateRepository(ctx, args.ProjectID, args.Location, args.RepositoryID, args.Format)
		if err != nil {
			if strings.Contains(err.Error(), "AlreadyExists") {
				res, err = arClient.GetRepository(ctx, args.ProjectID, args.Location, args.RepositoryID)
				if err != nil {
					return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to get existing repository: %w", err)
				}
			} else {
				return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to create repository: %w", err)
			}
		}

		if args.ServiceAccountEmail != "" {
			_, err := iamClient.AddIAMRoleBinding(ctx, args.ProjectID, "roles/artifactregistry.writer", fmt.Sprintf("serviceAccount:%s", args.ServiceAccountEmail))
			if err != nil {
				return &mcp.CallToolResult{}, nil, fmt.Errorf("repository created, but failed to grant permissions: %w", err)
			}
		}

		return &mcp.CallToolResult{}, res, nil
	}
	mcp.AddTool(server, &mcp.Tool{Name: "artifactregistry.setup_repository", Description: "Sets up a new Artifact Registry repository. Optionally, it can grant Artifact Registry Writer permissions to a service account."}, setupRepositoryToolFunc)
}
