// Copyright 2025 Google LLC
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

package bm25

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	bm25client "devops-mcp-server/bm25/client"
)

type Handler struct {
	BM25Client bm25client.BM25Client
}

// Register registers the rag tools with the MCP server.
func (h *Handler) Register(server *mcp.Server) {
	addQueryPatternTool(server, h.BM25Client)
	addQueryKnowledgeTool(server, h.BM25Client)
}

type QueryArgs struct {
	Query string `json:"query" jsonschema:"The query to search for."`
}


var queryPatternToolFunc func(ctx context.Context, req *mcp.CallToolRequest, args QueryArgs) (*mcp.CallToolResult, any, error)
var queryKnowledgeToolFunc func(ctx context.Context, req *mcp.CallToolRequest, args QueryArgs) (*mcp.CallToolResult, any, error)

func addQueryPatternTool(server *mcp.Server, bm25Client bm25client.BM25Client) {
	queryPatternToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args QueryArgs) (*mcp.CallToolResult, any, error) {
		res, err := bm25Client.QueryPatterns(ctx, args.Query)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to query patterns: %w", err)
		}
		return &mcp.CallToolResult{}, map[string]any{"cicd-patterns": res}, nil
	}
	mcp.AddTool(server, &mcp.Tool{Name: "bm25.search_common_cicd_patterns", Description: "Find common CICD patterns in the database."}, queryPatternToolFunc)
}

func addQueryKnowledgeTool(server *mcp.Server, bm25Client bm25client.BM25Client) {
	queryKnowledgeToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args QueryArgs) (*mcp.CallToolResult, any, error) {
		res, err := bm25Client.Queryknowledge(ctx, args.Query)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to query knowledge: %w", err)
		}
		return &mcp.CallToolResult{}, map[string]any{"knowledge": res}, nil
	}
	mcp.AddTool(server, &mcp.Tool{Name: "bm25.query_knowledge", Description: "Find knowledge snippets in the knowledge database."}, queryKnowledgeToolFunc)
}
