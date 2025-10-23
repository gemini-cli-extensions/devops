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

package artifactregistry

import (
	"context"
	"fmt"
	"log"

	"devops-mcp-server/artifactregistryiface"

	artifactregistry "cloud.google.com/go/artifactregistry/apiv1"
	artifactregistrypb "cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Client is an interface for interacting with the Artifact Registry API.
type Client interface {
	GetRepository(ctx context.Context, projectID, location, repositoryID string) (*artifactregistrypb.Repository, error)
	CreateRepository(ctx context.Context, projectID, location, repositoryID, format string) (*artifactregistrypb.Repository, error)
}

// client is a client for interacting with the Artifact Registry API.
type client struct {
	client artifactregistryiface.GRPClient
}

// NewClient creates a new Artifact Registry client.
func NewClient(c artifactregistryiface.GRPClient) (Client, error) {
	return &client{client: c}, nil
}

// GetRepository gets a repository from Artifact Registry.
func (c *client) GetRepository(ctx context.Context, projectID, location, repositoryID string) (*artifactregistrypb.Repository, error) {
	req := &artifactregistrypb.GetRepositoryRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/repositories/%s", projectID, location, repositoryID),
	}

	repo, err := c.client.GetRepository(ctx, req)
	return repo, err
}

// CreateRepository creates a new Artifact Registry repository.
func (c *client) CreateRepository(ctx context.Context, projectID, location, repositoryID, format string) (*artifactregistrypb.Repository, error) {
	// First, check if the repository already exists.
	repo, err := c.GetRepository(ctx, projectID, location, repositoryID)
	if err == nil {
		return repo, nil
	}

	s, ok := status.FromError(err)
	if !ok || s.Code() != codes.NotFound {
		return nil, fmt.Errorf("failed to get repository: %v", err)
	}

	req := &artifactregistrypb.CreateRepositoryRequest{
		Parent:       fmt.Sprintf("projects/%s/locations/%s", projectID, location),
		RepositoryId: repositoryID,
		Repository: &artifactregistrypb.Repository{
			Format: artifactregistrypb.Repository_Format(artifactregistrypb.Repository_Format_value[format]),
		},
	}

	op, err := c.client.CreateRepository(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %v", err)
	}

	repo, err = op.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for repository creation: %v", err)
	}

	return repo, nil
}

type createRepositoryOperation struct {
	*artifactregistry.CreateRepositoryOperation
}

func (op *createRepositoryOperation) Wait(ctx context.Context) (*artifactregistrypb.Repository, error) {
	return op.CreateRepositoryOperation.Wait(ctx)
}

// gRPCClient is a client for interacting with the Artifact Registry API.
type gRPCClient struct {
	*artifactregistry.Client
}

func (c *gRPCClient) GetRepository(ctx context.Context, req *artifactregistrypb.GetRepositoryRequest) (*artifactregistrypb.Repository, error) {
	return c.Client.GetRepository(ctx, req)
}

func (c *gRPCClient) CreateRepository(ctx context.Context, req *artifactregistrypb.CreateRepositoryRequest) (artifactregistryiface.CreateRepositoryOperation, error) {
	op, err := c.Client.CreateRepository(ctx, req)
	if err != nil {
		log.Printf("%v", err)
		return nil, err
	}
	return &createRepositoryOperation{op}, nil
}

// NewGRPCClient creates a new gRPC Artifact Registry client.
func NewGRPCClient(ctx context.Context) (artifactregistryiface.GRPClient, error) {
	c, err := artifactregistry.NewClient(ctx)
	if err != nil {
		log.Printf("%v", err)
		return nil, fmt.Errorf("failed to create artifact registry client: %v", err)
	}
	return &gRPCClient{c}, nil
}
