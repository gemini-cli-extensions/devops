package client

import (
	"context"
	"fmt"
	"time"

	devconnectv1 "google.golang.org/api/developerconnect/v1"
)

// DevConnectClient is an interface for interacting with the Developer Connect API.
type DevConnectClient interface {
	CreateConnection(ctx context.Context, projectID, location, connectionID string) (*devconnectv1.Connection, error)
	CreateGitRepositoryLink(ctx context.Context, projectID, location, connectionID, repoLinkID, repoURI string) (*devconnectv1.GitRepositoryLink, error)
	ListConnections(ctx context.Context, projectID, location string) (*ListResult[*devconnectv1.Connection], error)
	GetConnection(ctx context.Context, projectID, location, connectionID string) (*devconnectv1.Connection, error)
	FindGitRepositoryLinksForGitRepo(ctx context.Context, projectID, location, repoURI string) (*ListResult[*devconnectv1.GitRepositoryLink], error)
}

// ListResult defines a generic struct to wrap a list of items.

type ListResult[T any] struct {
	Items []T `json:"items"`
}

// Client is a client for interacting with the Developer Connect API.

type Client struct {
	service *devconnectv1.Service
}

// New creates a new DevConnectClient.
func New(ctx context.Context) (DevConnectClient, error) {
	service, err := devconnectv1.NewService(ctx)
	if err != nil {

		return nil, fmt.Errorf("failed to create developer connect service: %v", err)

	}
	return &Client{service: service}, nil
}

type clientKey struct{}

// WithClient returns a new context with the provided client.
func WithClient(ctx context.Context, client DevConnectClient) context.Context {
	return context.WithValue(ctx, clientKey{}, client)
}

// ClientFrom returns the client from the context.
func ClientFrom(ctx context.Context) (DevConnectClient, bool) {
	client, ok := ctx.Value(clientKey{}).(DevConnectClient)
	return client, ok
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
func (c *Client) ListConnections(ctx context.Context, projectID, location string) (*ListResult[*devconnectv1.Connection], error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)

	resp, err := c.service.Projects.Locations.Connections.List(parent).Do()

	if err != nil {

		return nil, fmt.Errorf("failed to list connections: %v", err)

	}

	return &ListResult[*devconnectv1.Connection]{Items: resp.Connections}, nil

}

// GetConnection gets a Developer Connect connection.
func (c *Client) GetConnection(ctx context.Context, projectID, location, connectionID string) (*devconnectv1.Connection, error) {
	name := fmt.Sprintf("projects/%s/locations/%s/connections/%s", projectID, location, connectionID)
	return c.service.Projects.Locations.Connections.Get(name).Do()
}

// FindGitRepositoryLinksForGitRepo finds already configured Developer Connect Git Repository Links for a particular git repository.
func (c *Client) FindGitRepositoryLinksForGitRepo(ctx context.Context, projectID, location, repoURI string) (*ListResult[*devconnectv1.GitRepositoryLink], error) {

	parent := fmt.Sprintf("projects/%s/locations/%s/connections/-", projectID, location)
	resp, err := c.service.Projects.Locations.Connections.GitRepositoryLinks.List(parent).Filter(fmt.Sprintf("clone_uri=\"%s\"", repoURI)).Do()

	if err != nil {
		return nil, fmt.Errorf("failed to list git repository links: %v", err)
	}

	return &ListResult[*devconnectv1.GitRepositoryLink]{Items: resp.GitRepositoryLinks}, nil
}
