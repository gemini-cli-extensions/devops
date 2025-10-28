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
	addListConnectionsTool(server, d)
	return nil
}

type ListConnectionsArgs struct {
	ProjectID string `json:"project_id" jsonschema:"The Google Cloud project ID."`
	Location  string `json:"location" jsonschema:"The Google Cloud location for the repository."`
}

var listConnectionsToolFunc func(ctx context.Context, req *mcp.CallToolRequest, args ListConnectionsArgs) (*mcp.CallToolResult, any, error)

func addListConnectionsTool(server *mcp.Server, dcClient devconnectclient.DeveloperConnectClient) {
	listConnectionsToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args ListConnectionsArgs) (*mcp.CallToolResult, any, error) {
		res, err := dcClient.ListConnections(ctx, args.ProjectID, args.Location)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to list connections: %w", err)
		}
		return &mcp.CallToolResult{}, res, nil
	}
	mcp.AddTool(server, &mcp.Tool{Name: "devconnect.list_connections", Description: "Lists Developer Connect connections."}, listConnectionsToolFunc)
}
