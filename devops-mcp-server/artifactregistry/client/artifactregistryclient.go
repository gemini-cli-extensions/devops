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

package artifactregistryclient

import (
	"context"
	"fmt"

	artifactregistry "cloud.google.com/go/artifactregistry/apiv1"
	artifactregistrypb "cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
)

// contextKey is a private type to use as a key for context values.
type contextKey string

const (
	artifactRegistryClientKey contextKey = "artifactRegistryClient"
)

// ClientFrom returns the ArtifactRegistryClient stored in the context, if any.
func ClientFrom(ctx context.Context) (ArtifactRegistryClient, bool) {
	client, ok := ctx.Value(artifactRegistryClientKey).(ArtifactRegistryClient)
	return client, ok
}

// ContextWithClient returns a new context with the provided ArtifactRegistryClient.
func ContextWithClient(ctx context.Context, client ArtifactRegistryClient) context.Context {
	return context.WithValue(ctx, artifactRegistryClientKey, client)
}

// ArtifactRegistryClient is an interface for interacting with the Artifact Registry API.
type ArtifactRegistryClient interface {
	GetRepository(ctx context.Context, projectID, location, repositoryID string) (*artifactregistrypb.Repository, error)
	CreateRepository(ctx context.Context, projectID, location, repositoryID, format string) (*artifactregistrypb.Repository, error)
	DeleteRepository(ctx context.Context, projectID, location, repositoryID string) error
}

// NewArtifactRegistryClient creates a new Artifact Registry client.
func NewArtifactRegistryClient(ctx context.Context) (ArtifactRegistryClient, error) {
	c, err := artifactregistry.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create artifact registry client: %v", err)
	}
	return &ArtifactRegistryClientImpl{v1client: c}, nil
}

// artifactRegistryClient is a client for interacting with the Artifact Registry API.
type ArtifactRegistryClientImpl struct {
	v1client *artifactregistry.Client
}

// GetRepository gets a repository from Artifact Registry.
func (c *ArtifactRegistryClientImpl) GetRepository(ctx context.Context, projectID, location, repositoryID string) (*artifactregistrypb.Repository, error) {
	req := &artifactregistrypb.GetRepositoryRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/repositories/%s", projectID, location, repositoryID),
	}

	repo, err := c.v1client.GetRepository(ctx, req)
	return repo, err
}

// CreateRepository creates a new Artifact Registry repository.
func (c *ArtifactRegistryClientImpl) CreateRepository(ctx context.Context, projectID, location, repositoryID, format string) (*artifactregistrypb.Repository, error) {

	req := &artifactregistrypb.CreateRepositoryRequest{
		Parent:       fmt.Sprintf("projects/%s/locations/%s", projectID, location),
		RepositoryId: repositoryID,
		Repository: &artifactregistrypb.Repository{
			Format: artifactregistrypb.Repository_Format(artifactregistrypb.Repository_Format_value[format]),
		},
	}

	op, err := c.v1client.CreateRepository(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %v", err)
	}

	repo, err := op.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for repository creation: %v", err)
	}

	return repo, nil
}

// DeleteRepository deletes an Artifact Registry repository.
func (c *ArtifactRegistryClientImpl) DeleteRepository(ctx context.Context, projectID, location, repositoryID string) error {
	req := &artifactregistrypb.DeleteRepositoryRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/repositories/%s", projectID, location, repositoryID),
	}

	op, err := c.v1client.DeleteRepository(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete repository: %v", err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for repository deletion: %v", err)
	}

	return nil
}
