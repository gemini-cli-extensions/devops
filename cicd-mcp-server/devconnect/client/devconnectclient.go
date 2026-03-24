// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package developerconnectclient

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/api/developerconnect/v1"
	"encoding/json"
)

// contextKey is a private type to use as a key for context values.
type contextKey string

const (
	// developerConnectClientKey is the private key used to store the DeveloperConnectClient in context.
	developerConnectClientKey contextKey = "developerConnectClient"
)

// ClientFrom returns the DeveloperConnectClient stored in the context, if any.
func ClientFrom(ctx context.Context) (DeveloperConnectClient, bool) {
	client, ok := ctx.Value(developerConnectClientKey).(DeveloperConnectClient)
	return client, ok
}

// ContextWithClient returns a new context with the provided DeveloperConnectClient.
func ContextWithClient(ctx context.Context, client DeveloperConnectClient) context.Context {
	return context.WithValue(ctx, developerConnectClientKey, client)
}

// DevConnectClient is an interface for interacting with the Developer Connect API.
type DeveloperConnectClient interface {
	GetConnection(ctx context.Context, projectID, location, connectionID string) (*developerconnect.Connection, error)
	CreateConnection(ctx context.Context, projectID, location, connectionID string) (*developerconnect.Connection, error)
	ListConnections(ctx context.Context, projectID, location string) ([]*developerconnect.Connection, error)
	CreateGitRepositoryLink(ctx context.Context, projectID, location, connectionID, repoLinkID, repoURI string) (*developerconnect.GitRepositoryLink, error)
	FindGitRepositoryLinksForGitRepo(ctx context.Context, projectID, location, repoURI string) ([]*developerconnect.GitRepositoryLink, error)
	GenerateUUID() string
}

// DeveloperConnectClientImpl is the concrete implementation.
type DeveloperConnectClientImpl struct {
	v1client *developerconnect.Service
}

// NewDeveloperConnectClient creates a new Developer Connect client.
func NewDeveloperConnectClient(ctx context.Context) (DeveloperConnectClient, error) {
	c, err := developerconnect.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create developer connect client: %v", err)
	}
	return &DeveloperConnectClientImpl{v1client: c}, nil
}

func (c *DeveloperConnectClientImpl) GenerateUUID() string {
	return uuid.New().String()
}

// CreateConnection creates a new Developer Connect connection.
func (c *DeveloperConnectClientImpl) CreateConnection(ctx context.Context, projectID, location, connectionID string) (*developerconnect.Connection, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
	conn := &developerconnect.Connection{
		GithubConfig: &developerconnect.GitHubConfig{
			GithubApp: "DEVELOPER_CONNECT",
		},
	}

	op, err := c.v1client.Projects.Locations.Connections.Create(parent, conn).ConnectionId(connectionID).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	var newConn developerconnect.Connection
	err = c.waitForOperation(ctx, op, &newConn)
	return &newConn, err
}

// CreateGitRepositoryLink creates a new Developer Connect Git Repository Link.
func (c *DeveloperConnectClientImpl) CreateGitRepositoryLink(ctx context.Context, projectID, location, connectionID, repoLinkID, repoURI string) (*developerconnect.GitRepositoryLink, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s/connections/%s", projectID, location, connectionID)
	link := &developerconnect.GitRepositoryLink{
		CloneUri: repoURI,
	}
	op, err := c.v1client.Projects.Locations.Connections.GitRepositoryLinks.Create(parent, link).GitRepositoryLinkId(repoLinkID).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	var newLink developerconnect.GitRepositoryLink
	err = c.waitForOperation(ctx, op, &newLink)
	return &newLink, err
}

// GetConnection gets a Developer Connect connection.
func (c *DeveloperConnectClientImpl) GetConnection(ctx context.Context, projectID, location, connectionID string) (*developerconnect.Connection, error) {
	name := fmt.Sprintf("projects/%s/locations/%s/connections/%s", projectID, location, connectionID)
	return c.v1client.Projects.Locations.Connections.Get(name).Context(ctx).Do()
}

// ListConnections lists Developer Connect connections.
func (c *DeveloperConnectClientImpl) ListConnections(ctx context.Context, projectID, location string) ([]*developerconnect.Connection, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
	resp, err := c.v1client.Projects.Locations.Connections.List(parent).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	return resp.Connections, nil
}

// FindGitRepositoryLinksForGitRepo finds already configured Developer Connect Git Repository Links for a particular git repository.
func (c *DeveloperConnectClientImpl) FindGitRepositoryLinksForGitRepo(ctx context.Context, projectID, location, repoURI string) ([]*developerconnect.GitRepositoryLink, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s/connections/-", projectID, location)
	filter := fmt.Sprintf("clone_uri=\"%s\"", repoURI)
	resp, err := c.v1client.Projects.Locations.Connections.GitRepositoryLinks.List(parent).Filter(filter).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	return resp.GitRepositoryLinks, nil
}

func (c *DeveloperConnectClientImpl) waitForOperation(ctx context.Context, op *developerconnect.Operation, out any) error {
	for !op.Done {
		time.Sleep(5 * time.Second)
		var err error
		op, err = c.v1client.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("failed to poll operation: %w", err)
		}
	}

	if op.Error != nil {
		return fmt.Errorf("operation failed: %s", op.Error.Message)
	}

	b, err := json.Marshal(op.Response)
	if err != nil {
		return fmt.Errorf("failed to marshal operation response: %w", err)
	}

	return json.Unmarshal(b, out)
}
