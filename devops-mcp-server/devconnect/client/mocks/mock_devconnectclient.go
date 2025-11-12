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

	"google.golang.org/api/developerconnect/v1"
)

// MockDevConnectClient is a mock implementation of the DevConnectClient interface.
type MockDevConnectClient struct {
	CreateConnectionFunc                 func(ctx context.Context, projectID, location, connectionID string) (*developerconnect.Connection, error)
	CreateGitRepositoryLinkFunc          func(ctx context.Context, projectID, location, connectionID, repoLinkID, repoURI string) (*developerconnect.GitRepositoryLink, error)
	ListConnectionsFunc                  func(ctx context.Context, projectID, location string) ([]*developerconnect.Connection, error)
	GetConnectionFunc                    func(ctx context.Context, projectID, location, connectionID string) (*developerconnect.Connection, error)
	FindGitRepositoryLinksForGitRepoFunc func(ctx context.Context, projectID, location, repoURI string) ([]*developerconnect.GitRepositoryLink, error)
}

// CreateConnection mocks the CreateConnection method.
func (m *MockDevConnectClient) CreateConnection(ctx context.Context, projectID, location, connectionID string) (*developerconnect.Connection, error) {
	return m.CreateConnectionFunc(ctx, projectID, location, connectionID)
}

// CreateGitRepositoryLink mocks the CreateGitRepositoryLink method.
func (m *MockDevConnectClient) CreateGitRepositoryLink(ctx context.Context, projectID, location, connectionID, repoLinkID, repoURI string) (*developerconnect.GitRepositoryLink, error) {
	return m.CreateGitRepositoryLinkFunc(ctx, projectID, location, connectionID, repoLinkID, repoURI)
}

// ListConnections mocks the ListConnections method.
func (m *MockDevConnectClient) ListConnections(ctx context.Context, projectID, location string) ([]*developerconnect.Connection, error) {
	return m.ListConnectionsFunc(ctx, projectID, location)
}

// GetConnection mocks the GetConnection method.
func (m *MockDevConnectClient) GetConnection(ctx context.Context, projectID, location, connectionID string) (*developerconnect.Connection, error) {
	return m.GetConnectionFunc(ctx, projectID, location, connectionID)
}

// FindGitRepositoryLinksForGitRepo mocks the FindGitRepositoryLinksForGitRepo method.
func (m *MockDevConnectClient) FindGitRepositoryLinksForGitRepo(ctx context.Context, projectID, location, repoURI string) ([]*developerconnect.GitRepositoryLink, error) {
	return m.FindGitRepositoryLinksForGitRepoFunc(ctx, projectID, location, repoURI)
}

// GenerateUUID generates a static UUID.
func (m *MockDevConnectClient) GenerateUUID() string {
	return "mock-connection"
}
