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

package devconnect

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	devconnectclient "devops-mcp-server/devconnect/client"
)

// AddTools adds all devconnect related tools to the mcp server.
// It expects the devconnectclient and mcp.Server to be in the context.
func AddTools(ctx context.Context, server *mcp.Server) error {
	d, ok := devconnectclient.ClientFrom(ctx)
	if !ok {
		return fmt.Errorf("devconnect client not found in context")
	}
	addSetupDevConnectConnectionTool(server, d)
	addAddDevConnectGitRepoLinkTool(server, d)
	return nil
}

type AddDevConnectGitRepoLinkArgs struct {
	ProjectID    string `json:"project_id" jsonschema:"The Google Cloud project ID."`
	Location     string `json:"location" jsonschema:"The Google Cloud location for the repository."`
	ConnectionID string `json:"connection_id" jsonschema:"The ID of the Developer Connect connection."`
	GitRepoURI   string `json:"git_repo_uri" jsonschema:"The URI of the git repository to link. e.g. https://github.com/gemini-cli-extensions/devops.git"`
}

var addDevConnectGitRepoLinkToolFunc func(ctx context.Context, req *mcp.CallToolRequest, args AddDevConnectGitRepoLinkArgs) (*mcp.CallToolResult, any, error)

func addAddDevConnectGitRepoLinkTool(server *mcp.Server, dcClient devconnectclient.DeveloperConnectClient) {
	addDevConnectGitRepoLinkToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args AddDevConnectGitRepoLinkArgs) (*mcp.CallToolResult, any, error) {
		// We need a repoLinkID. We can derive it from the URI.
		repoLinkID := "link-" + args.GitRepoURI
		newLink, err := dcClient.CreateGitRepositoryLink(ctx, args.ProjectID, args.Location, args.ConnectionID, repoLinkID, args.GitRepoURI)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to create git repository link: %w", err)
		}

		return &mcp.CallToolResult{}, newLink, nil
	}
	mcp.AddTool(server, &mcp.Tool{Name: "devconnect.add_git_repo_link", Description: "Creates a Developer Connect git repository link when a connection already exists."}, addDevConnectGitRepoLinkToolFunc)
}

type ResultWrapper struct {
	Message string
	Result  any
}

type SetupDevConnectConnectionArgs struct {
	ProjectID  string `json:"project_id" jsonschema:"The Google Cloud project ID."`
	Location   string `json:"location" jsonschema:"The Google Cloud location for the repository."`
	GitRepoURI string `json:"git_repo_uri" jsonschema:"The URI of the git repository to connect to."`
}

var setupDevConnectConnectionToolFunc func(ctx context.Context, req *mcp.CallToolRequest, args SetupDevConnectConnectionArgs) (*mcp.CallToolResult, any, error)

func addSetupDevConnectConnectionTool(server *mcp.Server, dcClient devconnectclient.DeveloperConnectClient) {
	setupDevConnectConnectionToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args SetupDevConnectConnectionArgs) (*mcp.CallToolResult, any, error) {
		// First, check if a git repository link already exists for this URI.
		existingLinks, err := dcClient.FindGitRepositoryLinksForGitRepo(ctx, args.ProjectID, args.Location, args.GitRepoURI)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to check for existing git repository links: %w", err)
		}
		if len(existingLinks) > 0 {
			return &mcp.CallToolResult{}, ResultWrapper{Message: "pre-exsisting connection found", Result: existingLinks[0]}, nil
		}

		newConnection, err := dcClient.CreateConnection(ctx, args.ProjectID, args.Location, dcClient.GenerateUUID())
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to create new connection: %w", err)
		}

		return &mcp.CallToolResult{}, ResultWrapper{Message: "Created connection, authorize the connection by visiting the `installationUri`. After authorizing, call the AddDevConnectGitRepoLink fo finalize.", Result: newConnection}, nil
	}
	mcp.AddTool(server, &mcp.Tool{Name: "devconnect.setup_connection", Description: "Sets up a Developer Connect connection and git repository link."}, setupDevConnectConnectionToolFunc)
}
