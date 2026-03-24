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

package mocks

import (
	"context"

	artifactregistrypb "cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
)

// MockArtifactRegistryClient is a mock of ArtifactRegistryClient interface.
type MockArtifactRegistryClient struct {
	GetRepositoryFunc    func(ctx context.Context, projectID, location, repositoryID string) (*artifactregistrypb.Repository, error)
	CreateRepositoryFunc func(ctx context.Context, projectID, location, repositoryID, format string) (*artifactregistrypb.Repository, error)
	DeleteRepositoryFunc func(ctx context.Context, projectID, location, repositoryID string) error
}

// DeleteRepository mocks the DeleteRepository method.
func (m *MockArtifactRegistryClient) DeleteRepository(ctx context.Context, projectID, location, repositoryID string) error {
	return m.DeleteRepositoryFunc(ctx, projectID, location, repositoryID)
}

// GetRepository mocks the GetRepository method.
func (m *MockArtifactRegistryClient) GetRepository(ctx context.Context, projectID, location, repositoryID string) (*artifactregistrypb.Repository, error) {
	return m.GetRepositoryFunc(ctx, projectID, location, repositoryID)
}

// CreateRepository mocks the CreateRepository method.
func (m *MockArtifactRegistryClient) CreateRepository(ctx context.Context, projectID, location, repositoryID, format string) (*artifactregistrypb.Repository, error) {
	return m.CreateRepositoryFunc(ctx, projectID, location, repositoryID, format)
}
