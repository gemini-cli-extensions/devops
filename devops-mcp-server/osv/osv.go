// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package osv

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	osvclient "devops-mcp-server/osv/client"
)

// Handler holds the clients for the osv service.
type Handler struct {
	OsvClient osvclient.OsvClient
}

// Register registers the osv tools with the MCP server.
func (h *Handler) Register(server *mcp.Server) {
	addScanSecretsTool(server, h.OsvClient)
}

type ScanSecretsArgs struct {
	Root string `json:"root" jsonschema:"The root directory to scan for secrets. Give the absolute directory path."`
}

var scanSecretsToolFunc func(ctx context.Context, req *mcp.CallToolRequest, args ScanSecretsArgs) (*mcp.CallToolResult, any, error)

func addScanSecretsTool(server *mcp.Server, oClient osvclient.OsvClient) {
	scanSecretsToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args ScanSecretsArgs) (*mcp.CallToolResult, any, error) {
		res, err := oClient.ScanSecrets(ctx, args.Root)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to scan for secrets: %w", err)
		}

		return &mcp.CallToolResult{}, map[string]any{"report": res}, nil
	}
	mcp.AddTool(server, &mcp.Tool{Name: "osv.scan_secrets", Description: "Scans the specified root directory for secrets using OSV."}, scanSecretsToolFunc)
}
