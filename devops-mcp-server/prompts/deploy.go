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

package prompts

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed deploy_prompt.md
var promptDeployText string

// Helps deploy applications to GCP.
func DeployPrompt(ctx context.Context, server *mcp.Server) {
	promptHandler := func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Description: "Helps deploy applications to GCP.",
			Messages: []*mcp.PromptMessage{
				{
					Role: "user",
					Content: &mcp.TextContent{
						Text: fmt.Sprintf(promptDeployText, req.Params.Arguments["query"]),
					},
				},
			},
		}, nil
	}

	// Create a server with a single prompt.
	prompt := &mcp.Prompt{
		Name:  "devops:deploy",
		Title: "Deploy an application to GCP.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "query",
				Description: "application to deploy, as explained by the user",
				Required:    true,
			},
		},
	}
	server.AddPrompt(prompt, promptHandler)
}
