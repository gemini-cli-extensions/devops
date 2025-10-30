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

package cloudrunclient

import (
	"context"
	"fmt"
	"os/exec"

	cloudrun "cloud.google.com/go/run/apiv2"
	cloudrunpb "cloud.google.com/go/run/apiv2/runpb"
)

// contextKey is a private type to use as a key for context values.
type contextKey string

const (
	cloudRunClientKey contextKey = "cloudRunClient"
)

// ClientFrom returns the CloudRunClient stored in the context, if any.
func ClientFrom(ctx context.Context) (CloudRunClient, bool) {
	client, ok := ctx.Value(cloudRunClientKey).(CloudRunClient)
	return client, ok
}

// ContextWithClient returns a new context with the provided CloudRunClient.
func ContextWithClient(ctx context.Context, client CloudRunClient) context.Context {
	return context.WithValue(ctx, cloudRunClientKey, client)
}

// CloudRunClient is an interface for interacting with the Cloud Run API.
type CloudRunClient interface {
	GetService(ctx context.Context, projectID, location, serviceName string) (*cloudrunpb.Service, error)
	CreateService(ctx context.Context, projectID, location, serviceName, imageURL string, port int32) (*cloudrunpb.Service, error)
	UpdateService(ctx context.Context, projectID, location, serviceName, imageURL, revisionName string, port int32, service *cloudrunpb.Service) (*cloudrunpb.Service, error)
	GetRevision(ctx context.Context, service *cloudrunpb.Service) (*cloudrunpb.Revision, error)
	DeployFromSource(ctx context.Context, projectID, location, serviceName, source string, port int32) error
}

// NewCloudRunClient creates a new CloudRunClient.
func NewCloudRunClient(ctx context.Context) (CloudRunClient, error) {
	servicesClient, err := cloudrun.NewServicesClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud run services client: %w", err)
	}
	revisionsClient, err := cloudrun.NewRevisionsClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud run revisions client: %w", err)
	}
	return &CloudRunClientImpl{servicesClient: servicesClient, revisionsClient: revisionsClient, execer: defaultExecer}, nil
}

// CloudRunClientImpl is a client for interacting with the Cloud Run API.
type CloudRunClientImpl struct {
	servicesClient  *cloudrun.ServicesClient
	revisionsClient *cloudrun.RevisionsClient
	execer          Exec
}

// Exec interface for running commands.
type Exec interface {
	Command(name string, arg ...string) *exec.Cmd
}

type execer struct{}

func (e *execer) Command(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

var defaultExecer Exec = &execer{}

// CreateService creates a new Cloud Run service.
func (c *CloudRunClientImpl) CreateService(ctx context.Context, projectID, location, serviceName, imageURL string, port int32) (*cloudrunpb.Service, error) {
	req := &cloudrunpb.CreateServiceRequest{
		Parent:    fmt.Sprintf("projects/%s/locations/%s", projectID, location),
		ServiceId: serviceName,
		Service: &cloudrunpb.Service{
			Template: &cloudrunpb.RevisionTemplate{
				Containers: []*cloudrunpb.Container{
					{
						Image: imageURL,
						Ports: []*cloudrunpb.ContainerPort{
							{
								ContainerPort: port,
							},
						},
					},
				},
			},
		},
	}
	op, err := c.servicesClient.CreateService(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}
	return op.Wait(ctx)
}

func (c *CloudRunClientImpl) GetService(ctx context.Context, projectID, location, serviceName string) (*cloudrunpb.Service, error) {
	servicePath := fmt.Sprintf("projects/%s/locations/%s/services/%s", projectID, location, serviceName)

	service, err := c.servicesClient.GetService(ctx, &cloudrunpb.GetServiceRequest{Name: servicePath})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}
	return service, nil
}

func (c *CloudRunClientImpl) GetRevision(ctx context.Context, service *cloudrunpb.Service) (*cloudrunpb.Revision, error) {
	// Get the latest revision
	latestRevision, err := c.revisionsClient.GetRevision(ctx, &cloudrunpb.GetRevisionRequest{Name: service.LatestReadyRevision})
	if err != nil {
		return nil, fmt.Errorf("failed to get latest revision: %w", err)
	}
	return latestRevision, nil
}

// UpdateService updates a service by creating a new Cloud Run revision with a new Docker image.
func (c *CloudRunClientImpl) UpdateService(ctx context.Context, projectID, location, serviceName, imageURL, revisionName string, port int32, service *cloudrunpb.Service) (*cloudrunpb.Service, error) {
	servicePath := fmt.Sprintf("projects/%s/locations/%s/services/%s", projectID, location, serviceName)

	// Create a new revision template based on the current service's template
	newTemplate := service.Template
	newTemplate.Containers[0].Image = imageURL
	if revisionName != "" {
		newTemplate.Revision = revisionName
	}
	newTemplate.Containers[0].Ports = []*cloudrunpb.ContainerPort{
		{
			ContainerPort: port,
		},
	}

	// Update the service with the new revision template
	updatedService := &cloudrunpb.Service{
		Name:     servicePath,
		Template: newTemplate,
	}

	op, err := c.servicesClient.UpdateService(ctx, &cloudrunpb.UpdateServiceRequest{Service: updatedService})
	if err != nil {
		return nil, fmt.Errorf("failed to update service: %w", err)
	}
	return op.Wait(ctx)
}

// DeployFromSource creates a new Cloud Run service or updates an existing one from source.
func (c *CloudRunClientImpl) DeployFromSource(ctx context.Context, projectID, location, serviceName, source string, port int32) error {
	args := []string{"run", "deploy", serviceName, "--project", projectID, "--region", location, "--source", source, "--format", "json", "--quiet"}
	if port != 0 {
		args = append(args, "--port", fmt.Sprintf("%d", port))
	}
	cmd := c.execer.Command("gcloud", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to deploy from source: %w, output: %s", err, out)
	}
	return nil
}
