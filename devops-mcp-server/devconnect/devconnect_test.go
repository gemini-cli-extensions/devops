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

package devconnect

import (
	"context"
	"testing"

	"devops-mcp-server/devconnect/client/mocks"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/developerconnect/v1"
)

func TestSetupDevConnectConnection_ExistingLink(t *testing.T) {
	mockClient := &mocks.MockDevConnectClient{
		FindGitRepositoryLinksForGitRepoFunc: func(ctx context.Context, projectID, location, repoURI string) ([]*developerconnect.GitRepositoryLink, error) {
			return []*developerconnect.GitRepositoryLink{
				{Name: "existing-link"},
			}, nil
		},
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "test"}, &mcp.ServerOptions{})
	addSetupDevConnectConnectionTool(server, mockClient)

	args := SetupDevConnectConnectionArgs{
		ProjectID:  "test-project",
		Location:   "us-central1",
		GitRepoURI: "https://github.com/test/repo.git",
	}

	_, res, err := setupDevConnectConnectionToolFunc(context.Background(), nil, args)
	if err != nil {
		t.Fatalf("tool function returned an error: %v", err)
	}

	resultWrapper, ok := res.(ResultWrapper)
	if !ok {
		t.Fatalf("result is not of the expected type")
	}

	link, ok := resultWrapper.Result.(*developerconnect.GitRepositoryLink)
	if !ok {
		t.Fatalf("result is not of the expected type for links")
	}

	if link.Name != "existing-link" {
		t.Errorf("expected link name 'existing-link', got '%s'", link.Name)
	}
}

func TestSetupDevConnectConnection_NoExistingLink_NoExistingConnection(t *testing.T) {
	mockClient := &mocks.MockDevConnectClient{
		FindGitRepositoryLinksForGitRepoFunc: func(ctx context.Context, projectID, location, repoURI string) ([]*developerconnect.GitRepositoryLink, error) {
			return []*developerconnect.GitRepositoryLink{}, nil
		},
		CreateConnectionFunc: func(ctx context.Context, projectID, location, connectionID string) (*developerconnect.Connection, error) {
			return &developerconnect.Connection{Name: "projects/test-project/locations/us-central1/connections/mcp-connection"}, nil
		},
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "test"}, &mcp.ServerOptions{})
	addSetupDevConnectConnectionTool(server, mockClient)

	args := SetupDevConnectConnectionArgs{
		ProjectID:  "test-project",
		Location:   "us-central1",
		GitRepoURI: "https://github.com/test/repo.git",
	}

	_, res, err := setupDevConnectConnectionToolFunc(context.Background(), nil, args)
	if err != nil {
		t.Fatalf("tool function returned an error: %v", err)
	}

	resultWrapper, ok := res.(ResultWrapper)
	if !ok {
		t.Fatalf("result is not of the expected type")
	}

	connection, ok := resultWrapper.Result.(*developerconnect.Connection)
	if !ok {
		t.Fatalf("result is not of the expected type for connections")
	}

	if connection.Name != "projects/test-project/locations/us-central1/connections/mcp-connection" {
		t.Errorf("expected link name 'projects/test-project/locations/us-central1/connections/mcp-connection', got '%s'", connection.Name)
	}
}

func TestAddDevConnectGitRepoLink(t *testing.T) {
	mockClient := &mocks.MockDevConnectClient{
		CreateGitRepositoryLinkFunc: func(ctx context.Context, projectID, location, connectionID, repoLinkID, repoURI string) (*developerconnect.GitRepositoryLink, error) {
			return &developerconnect.GitRepositoryLink{Name: "new-link"}, nil
		},
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "test"}, &mcp.ServerOptions{})
	addAddDevConnectGitRepoLinkTool(server, mockClient)

	args := AddDevConnectGitRepoLinkArgs{
		ProjectID:    "test-project",
		Location:     "us-central1",
		ConnectionID: "test-connection",
		GitRepoURI:   "https://github.com/test/repo.git",
	}

	_, result, err := addDevConnectGitRepoLinkToolFunc(context.Background(), nil, args)
	if err != nil {
		t.Fatalf("tool function returned an error: %v", err)
	}

	link, ok := result.(*developerconnect.GitRepositoryLink)
	if !ok {
		t.Fatalf("tool function returned incorrect type")
	}

	if link.Name != "new-link" {
		t.Errorf("expected link name 'new-link', got '%s'", link.Name)
	}
}
