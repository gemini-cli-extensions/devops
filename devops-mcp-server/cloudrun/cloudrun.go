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

package cloudrun

import (
	"context"
	"fmt"
	"os/exec"

	run "cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
)

// Exec interface for running commands.
type Exec interface {
	Command(name string, arg ...string) *exec.Cmd
}

type execer struct{}

func (e *execer) Command(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

var defaultExecer Exec = &execer{}


// Client is a client for interacting with the Cloud Run API.
type Client struct {
	servicesClient  *run.ServicesClient
	revisionsClient *run.RevisionsClient
	execer          Exec
}

// NewClient creates a new Client.
func NewClient(ctx context.Context) (*Client, error) {
	servicesClient, err := run.NewServicesClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud run services client: %v", err)
	}
	revisionsClient, err := run.NewRevisionsClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud run revisions client: %v", err)
	}
	return &Client{servicesClient: servicesClient, revisionsClient: revisionsClient, execer: defaultExecer}, nil
}

// CreateService creates a new Cloud Run service.
func (c *Client) CreateService(ctx context.Context, projectID, location, serviceName, imageURL string, port int32) (*runpb.Service, error) {
	req := &runpb.CreateServiceRequest{
		Parent:    fmt.Sprintf("projects/%s/locations/%s", projectID, location),
		ServiceId: serviceName,
		Service: &runpb.Service{
			Template: &runpb.RevisionTemplate{
				Containers: []*runpb.Container{
					{
						Image: imageURL,
						Ports: []*runpb.ContainerPort{
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
		return nil, fmt.Errorf("failed to create service: %v", err)
	}
	return op.Wait(ctx)
}

// CreateRevision creates a new Cloud Run revision for a service with a new Docker image.
func (c *Client) CreateRevision(ctx context.Context, projectID, location, serviceName, imageURL, revisionName string) (*runpb.Revision, error) {
	servicePath := fmt.Sprintf("projects/%s/locations/%s/services/%s", projectID, location, serviceName)

	// Get the current service to get its template
	service, err := c.servicesClient.GetService(ctx, &runpb.GetServiceRequest{Name: servicePath})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %v", err)
	}

	// Create a new revision template based on the current service's template
	newTemplate := service.Template
	newTemplate.Containers[0].Image = imageURL
	if revisionName != "" {
		newTemplate.Revision = revisionName
	}

	// Update the service with the new revision template
	updatedService := &runpb.Service{
		Name:     servicePath,
		Template: newTemplate,
	}

	op, err := c.servicesClient.UpdateService(ctx, &runpb.UpdateServiceRequest{Service: updatedService})
	if err != nil {
		return nil, fmt.Errorf("failed to update service: %v", err)
	}
	updatedService, err = op.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for service update: %v", err)
	}

	// Get the latest revision
	latestRevision, err := c.revisionsClient.GetRevision(ctx, &runpb.GetRevisionRequest{Name: updatedService.LatestReadyRevision})
	if err != nil {
		return nil, fmt.Errorf("failed to get latest revision: %v", err)
	}

	return latestRevision, nil
}

// DeployFromImage deploys a new Cloud Run service from a container image.
func (c *Client) DeployFromImage(ctx context.Context, projectID, location, serviceName, imageURL string, port int32) error {
	args := []string{"run", "deploy", serviceName, "--image", imageURL, "--project", projectID, "--region", location, "--format", "json", "--quiet"}
	if port != 0 {
		args = append(args, "--port", fmt.Sprintf("%d", port))
	}
	cmd := c.execer.Command("gcloud", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to deploy from image: %v, output: %s", err, out)
	}

	return nil
}

// DeployFromSource deploys a new Cloud Run service or updates an existing one from source.
func (c *Client) DeployFromSource(ctx context.Context, projectID, location, serviceName, source string, port int32) error {
	args := []string{"run", "deploy", serviceName, "--project", projectID, "--region", location, "--source", source, "--format", "json", "--quiet"}
	if port != 0 {
		args = append(args, "--port", fmt.Sprintf("%d", port))
	}
	cmd := c.execer.Command("gcloud", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to deploy from source: %v, output: %s", err, out)
	}
	return nil
}