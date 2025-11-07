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

package cloudrun

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cloudrunclient "devops-mcp-server/cloudrun/client"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AddTools adds all cloud run related tools to the mcp server.
func AddTools(ctx context.Context, server *mcp.Server) error {
	c, ok := cloudrunclient.ClientFrom(ctx)
	if !ok {
		return fmt.Errorf("cloud run client not found in context")
	}

	addListServicesTool(server, c)
	addCreateServiceTool(server, c)
	addCreateServiceFromSourceTool(server, c)
	return nil
}

type ListServicesArgs struct {
	ProjectID    string `json:"project_id" jsonschema:"The Google Cloud project ID."`
	Location     string `json:"location" jsonschema:"The Google Cloud location."`
}

var listServicesToolFunc func(ctx context.Context, req *mcp.CallToolRequest, args ListServicesArgs) (*mcp.CallToolResult, any, error)

func addListServicesTool(server *mcp.Server, crClient cloudrunclient.CloudRunClient) {
	listServicesToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args ListServicesArgs) (*mcp.CallToolResult, any, error) {
		services, err := crClient.ListServices(ctx, args.ProjectID, args.Location)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to list services: %w", err)
		}
		return &mcp.CallToolResult{}, services, nil
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudrun.list_services", Description: "Lists the Cloud Run service in a specified GCP project and location."}, listServicesToolFunc)

}

type CreateServiceArgs struct {
	ProjectID    string `json:"project_id" jsonschema:"The Google Cloud project ID."`
	Location     string `json:"location" jsonschema:"The Google Cloud location."`
	ServiceName  string `json:"service_name" jsonschema:"The name of the Cloud Run service."`
	RevisionName string `json:"revision_name" jsonschema:"The name of the Cloud run revision."`
	ImageURL     string `json:"image_url" jsonschema:"The URL of the container image to deploy."`
	Port         int32  `json:"port,omitempty" jsonschema:"The port the container listens on."`
}

var createServiceToolFunc func(ctx context.Context, req *mcp.CallToolRequest, args CreateServiceArgs) (*mcp.CallToolResult, any, error)

func addCreateServiceTool(server *mcp.Server, crClient cloudrunclient.CloudRunClient) {
	createServiceToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args CreateServiceArgs) (*mcp.CallToolResult, any, error) {
		// Attempt to create the service
		service, err := crClient.CreateService(ctx, args.ProjectID, args.Location, args.ServiceName, args.ImageURL, args.Port)
		if err == nil {
			return &mcp.CallToolResult{}, service, nil
		}

		// Check if the error was "AlreadyExists"
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.AlreadyExists {
			// This was some other error
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to create service: %w", err)
		}

		// Service already exists, so we must update it.
		service, err = crClient.GetService(ctx, args.ProjectID, args.Location, args.ServiceName)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to get service: %w", err)
		}
		service, err = crClient.UpdateService(ctx, args.ProjectID, args.Location, args.ServiceName, args.ImageURL, args.RevisionName, args.Port, service)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to update service with new revision: %w", err)
		}
		revision, err := crClient.GetRevision(ctx, service)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to get revision: %w", err)
		}
		return &mcp.CallToolResult{}, revision, nil
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudrun.create_service", Description: "Creates a new Cloud Run service from a container image. This tool may take a couple minutes to finish running."}, createServiceToolFunc)
}

type CreateServiceFromSourceArgs struct {
	ProjectID   string `json:"project_id" jsonschema:"The Google Cloud project ID."`
	Location    string `json:"location" jsonschema:"The Google Cloud location."`
	ServiceName string `json:"service_name" jsonschema:"The name of the Cloud Run service."`
	Source      string `json:"source" jsonschema:"The path to the source code to deploy."`
	Port        int32  `json:"port,omitempty" jsonschema:"The port the container listens on."`
}

var createServiceFromSourceToolFunc func(ctx context.Context, req *mcp.CallToolRequest, args CreateServiceFromSourceArgs) (*mcp.CallToolResult, any, error)

func addCreateServiceFromSourceTool(server *mcp.Server, crClient cloudrunclient.CloudRunClient) {
	createServiceFromSourceToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args CreateServiceFromSourceArgs) (*mcp.CallToolResult, any, error) {
		err := crClient.DeployFromSource(ctx, args.ProjectID, args.Location, args.ServiceName, args.Source, args.Port)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to create service: %w", err)
		}
		service, err := crClient.GetService(ctx, args.ProjectID, args.Location, args.ServiceName)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to get service: %w", err)
		}
		return &mcp.CallToolResult{}, service, nil
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudrun.create_service_from_source", Description: "Creates a new Cloud Run service or updates an existing one from source. This tool may take a couple minutes to finish running."}, createServiceFromSourceToolFunc)
}
