// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rag

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	ragclient "devops-mcp-server/rag/client"
)

// Handler holds the clients for the rag service.
type Handler struct {
	RagClient ragclient.RagClient
}

// Register registers the rag tools with the MCP server.
func (h *Handler) Register(server *mcp.Server) {
	addQueryPatternTool(server, h.RagClient)
	addQueryKnowledgeTool(server, h.RagClient)
}

type QueryArgs struct {
	Query string `json:"query" jsonschema:"The query to search for."`
}

var queryPatternToolFunc func(ctx context.Context, req *mcp.CallToolRequest, args QueryArgs) (*mcp.CallToolResult, any, error)
var queryKnowledgeToolFunc func(ctx context.Context, req *mcp.CallToolRequest, args QueryArgs) (*mcp.CallToolResult, any, error)

func addQueryPatternTool(server *mcp.Server, ragClient ragclient.RagClient) {
	queryPatternToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args QueryArgs) (*mcp.CallToolResult, any, error) {
		res, err := ragClient.QueryPatterns(ctx, args.Query)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to query patterns: %w", err)
		}
		return &mcp.CallToolResult{}, map[string]any{"cicd-patterns": res}, nil
	}
	mcp.AddTool(server, &mcp.Tool{Name: "rag.search_common_cicd_patterns", Description: "Find common CICD patterns in the database."}, queryPatternToolFunc)
}

func addQueryKnowledgeTool(server *mcp.Server, ragClient ragclient.RagClient) {
	queryKnowledgeToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args QueryArgs) (*mcp.CallToolResult, any, error) {
		res, err := ragClient.Queryknowledge(ctx, args.Query)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to query knowledge: %w", err)
		}
		return &mcp.CallToolResult{}, map[string]any{"knowledge": res}, nil
	}
	mcp.AddTool(server, &mcp.Tool{Name: "rag.query_knowledge", Description: "Find knowledge snippets in the knowledge database."}, queryKnowledgeToolFunc)
}
