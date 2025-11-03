package developerconnectclient

import (
	"context"
	"fmt"

	devconnect "cloud.google.com/go/developerconnect/apiv1"
	devconnectpb "cloud.google.com/go/developerconnect/apiv1/developerconnectpb"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
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
	GetConnection(ctx context.Context, projectID, location, connectionID string) (*devconnectpb.Connection, error)
	CreateConnection(ctx context.Context, projectID, location, connectionID string) (*devconnectpb.Connection, error)
	ListConnections(ctx context.Context, projectID, location string) ([]*devconnectpb.Connection, error)
	CreateGitRepositoryLink(ctx context.Context, projectID, location, connectionID, repoLinkID, repoURI string) (*devconnectpb.GitRepositoryLink, error)
	FindGitRepositoryLinksForGitRepo(ctx context.Context, projectID, location, repoURI string) ([]*devconnectpb.GitRepositoryLink, error)
	GenerateUUID() string
}

// DeveloperConnectClientImpl is the concrete implementation.
type DeveloperConnectClientImpl struct {
	v1client *devconnect.Client
}

// NewDeveloperConnectClient creates a new Developer Connect client.
func NewDeveloperConnectClient(ctx context.Context) (DeveloperConnectClient, error) {
	c, err := devconnect.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create developer connect client: %v", err)
	}
	return &DeveloperConnectClientImpl{v1client: c}, nil
}

func (c *DeveloperConnectClientImpl) GenerateUUID() string {
	return uuid.New().String()
}

// CreateConnection creates a new Developer Connect connection.
func (c *DeveloperConnectClientImpl) CreateConnection(ctx context.Context, projectID, location, connectionID string) (*devconnectpb.Connection, error) {
	req := &devconnectpb.CreateConnectionRequest{
		Parent:       fmt.Sprintf("projects/%s/locations/%s", projectID, location),
		ConnectionId: connectionID,
		Connection: &devconnectpb.Connection{
			ConnectionConfig: &devconnectpb.Connection_GithubConfig{
				GithubConfig: &devconnectpb.GitHubConfig{
					GithubApp: devconnectpb.GitHubConfig_DEVELOPER_CONNECT,
				},
			},
		},
	}

	// This returns a CreateConnectionOperation object (the LRO wrapper)
	op, err := c.v1client.CreateConnection(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start connection creation: %v", err)
	}

	conn, err := op.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection after waiting: %v", err)
	}

	return conn, nil
}

// CreateGitRepositoryLink creates a new Developer Connect Git Repository Link.
func (c *DeveloperConnectClientImpl) CreateGitRepositoryLink(ctx context.Context, projectID, location, connectionID, repoLinkID, repoURI string) (*devconnectpb.GitRepositoryLink, error) {
	req := &devconnectpb.CreateGitRepositoryLinkRequest{
		Parent:              fmt.Sprintf("projects/%s/locations/%s/connections/%s", projectID, location, connectionID),
		GitRepositoryLinkId: repoLinkID,
		GitRepositoryLink: &devconnectpb.GitRepositoryLink{
			CloneUri: repoURI,
		},
	}

	// This returns a CreateGitRepositoryLinkOperation object (the LRO wrapper)
	op, err := c.v1client.CreateGitRepositoryLink(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start Git Repository Link creation: %v", err)
	}

	// Use the built-in .Wait() method for polling
	link, err := op.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Git Repository Link after waiting: %v", err)
	}

	return link, nil
}

// GetConnection gets a Developer Connect connection.
func (c *DeveloperConnectClientImpl) GetConnection(ctx context.Context, projectID, location, connectionID string) (*devconnectpb.Connection, error) {
	name := fmt.Sprintf("projects/%s/locations/%s/connections/%s", projectID, location, connectionID)
	req := &devconnectpb.GetConnectionRequest{
		Name: name,
	}
	return c.v1client.GetConnection(ctx, req)
}

// ListConnections lists Developer Connect connections.
func (c *DeveloperConnectClientImpl) ListConnections(ctx context.Context, projectID, location string) ([]*devconnectpb.Connection, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
	req := &devconnectpb.ListConnectionsRequest{
		Parent: parent,
	}

	// Using an iterator from the modern Go client library
	it := c.v1client.ListConnections(ctx, req)
	var connections []*devconnectpb.Connection
	for {
		conn, err := it.Next()
		if err == iterator.Done { // Assuming you use the iterator package
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate connections: %v", err)
		}
		connections = append(connections, conn)
	}

	return connections, nil
}

// FindGitRepositoryLinksForGitRepo finds already configured Developer Connect Git Repository Links for a particular git repository.
func (c *DeveloperConnectClientImpl) FindGitRepositoryLinksForGitRepo(ctx context.Context, projectID, location, repoURI string) ([]*devconnectpb.GitRepositoryLink, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s/connections/-", projectID, location)
	filter := fmt.Sprintf("clone_uri=\"%s\"", repoURI)
	req := &devconnectpb.ListGitRepositoryLinksRequest{
		Parent: parent,
		Filter: filter,
	}

	it := c.v1client.ListGitRepositoryLinks(ctx, req)
	var links []*devconnectpb.GitRepositoryLink
	for {
		link, err := it.Next()
		if err == iterator.Done { // Assuming you use the iterator package
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate git repository links: %v", err)
		}
		links = append(links, link)
	}
	return links, nil
}
