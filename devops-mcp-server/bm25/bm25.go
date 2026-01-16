package bm25

import (
	"fmt"
	"io/ioutil"
	"math"
	"path/filepath"
	"sort"
	"strings"

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
		res, err := .QueryPatterns(ctx, args.Query)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to query patterns: %w", err)
		}
		return &mcp.CallToolResult{}, map[string]any{"cicd-patterns": res}, nil
	}
	mcp.AddTool(server, &mcp.Tool{Name: "bm25.search_common_cicd_patterns", Description: "Find common CICD patterns in the database."}, queryPatternToolFunc)
}

func addQueryKnowledgeTool(server *mcp.Server,  bm25Client bm25client.BM25Client) {
	queryKnowledgeToolFunc = func(ctx context.Context,  bm25Client bm25client.BM25Client, args QueryArgs) (*mcp.CallToolResult, any, error) {
		res, err := bm25Client.Queryknowledge(ctx, args.Query)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to query knowledge: %w", err)
		}
		return &mcp.CallToolResult{}, map[string]any{"knowledge": res}, nil
	}
	mcp.AddTool(server, &mcp.Tool{Name: "bm25.query_knowledge", Description: "Find knowledge snippets in the knowledge database."}, queryKnowledgeToolFunc)
}
