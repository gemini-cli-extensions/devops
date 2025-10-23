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
	"fmt"
	"time"

	devconnectv1 "google.golang.org/api/developerconnect/v1"
)

// Client is a client for interacting with the Developer Connect API.
type Client struct {
	service *devconnectv1.Service
}

// NewClient creates a new Client.
func NewClient(ctx context.Context) (*Client, error) {
	service, err := devconnectv1.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create developer connect service: %v", err)
	}
	return &Client{service: service}, nil
}

func (c *Client) waitForOperation(ctx context.Context, operation *devconnectv1.Operation) (*devconnectv1.Operation, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	for !operation.Done {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out waiting for operation: %v", ctx.Err())
		case <-time.After(1 * time.Second):
			op, err := c.service.Projects.Locations.Operations.Get(operation.Name).Do()
			if err != nil {
				return nil, fmt.Errorf("failed to get operation: %v", err)
			}
			operation = op
		}
	}
	return operation, nil
}

// CreateConnection creates a new Developer Connect connection.
func (c *Client) CreateConnection(ctx context.Context, projectID, location, connectionID string) (*devconnectv1.Connection, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
	req := &devconnectv1.Connection{
		GithubConfig: &devconnectv1.GitHubConfig{
			GithubApp: "DEVELOPER_CONNECT",
		},
	}
	op, err := c.service.Projects.Locations.Connections.Create(parent, req).ConnectionId(connectionID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %v", err)
	}
	op, err = c.waitForOperation(ctx, op)
	if err != nil {
		return nil, err
	}
	if op.Error != nil {
		return nil, fmt.Errorf("operation failed: %v", op.Error)
	}

	name := fmt.Sprintf("projects/%s/locations/%s/connections/%s", projectID, location, connectionID)
	return c.service.Projects.Locations.Connections.Get(name).Do()
}

// CreateGitRepositoryLink creates a new Developer Connect Git Repository Link.
func (c *Client) CreateGitRepositoryLink(ctx context.Context, projectID, location, connectionID, repoLinkID, repoURI string) (*devconnectv1.GitRepositoryLink, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s/connections/%s", projectID, location, connectionID)
	req := &devconnectv1.GitRepositoryLink{
		CloneUri: repoURI,
	}
	op, err := c.service.Projects.Locations.Connections.GitRepositoryLinks.Create(parent, req).GitRepositoryLinkId(repoLinkID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create git repository link: %v", err)
	}
	op, err = c.waitForOperation(ctx, op)
	if err != nil {
		return nil, err
	}
	if op.Error != nil {
		return nil, fmt.Errorf("operation failed: %v", op.Error)
	}

	name := fmt.Sprintf("%s/gitRepositoryLinks/%s", parent, repoLinkID)
	return c.service.Projects.Locations.Connections.GitRepositoryLinks.Get(name).Do()
}

// ListConnections lists Developer Connect connections.
func (c *Client) ListConnections(ctx context.Context, projectID, location string) ([]*devconnectv1.Connection, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
	resp, err := c.service.Projects.Locations.Connections.List(parent).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list connections: %v", err)
	}
	return resp.Connections, nil
}

// GetConnection gets a Developer Connect connection.
func (c *Client) GetConnection(ctx context.Context, projectID, location, connectionID string) (*devconnectv1.Connection, error) {
	name := fmt.Sprintf("projects/%s/locations/%s/connections/%s", projectID, location, connectionID)
	return c.service.Projects.Locations.Connections.Get(name).Do()
}

// FindGitRepositoryLinksForGitRepo finds already configured Developer Connect Git Repository Links for a particular git repository.
func (c *Client) FindGitRepositoryLinksForGitRepo(ctx context.Context, projectID, location, repoURI string) ([]*devconnectv1.GitRepositoryLink, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s/connections/-", projectID, location)
	resp, err := c.service.Projects.Locations.Connections.GitRepositoryLinks.List(parent).Filter(fmt.Sprintf("clone_uri=\"%s\"", repoURI)).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list git repository links: %v", err)
	}
	return resp.GitRepositoryLinks, nil
}
