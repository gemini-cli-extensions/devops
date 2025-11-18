// Copyright 2024 Google LLC
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

// AddTools adds all rag related tools to the mcp server.
// It expects the ragclient and mcp.Server to be in the context.
func AddTools(ctx context.Context, server *mcp.Server) error {
	r, ok := ragclient.ClientFrom(ctx)
	if !ok {
		return fmt.Errorf("rag client not found in context")
	}
	addQueryPatternTool(server, r)
	addQueryKnowledgeTool(server, r)
	return nil
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
		return &mcp.CallToolResult{}, res, nil
	}
	mcp.AddTool(server, &mcp.Tool{Name: "rag.query_pattern", Description: "Queries the RAG pattern database."}, queryPatternToolFunc)
}

func addQueryKnowledgeTool(server *mcp.Server, ragClient ragclient.RagClient) {
	queryKnowledgeToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args QueryArgs) (*mcp.CallToolResult, any, error) {
		res, err := ragClient.Queryknowledge(ctx, args.Query)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to query knowledge: %w", err)
		}
		return &mcp.CallToolResult{}, res, nil
	}
	mcp.AddTool(server, &mcp.Tool{Name: "rag.query_knowledge", Description: "Queries the RAG knowledge database."}, queryKnowledgeToolFunc)
}