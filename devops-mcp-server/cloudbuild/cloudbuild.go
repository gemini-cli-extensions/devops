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

package cloudbuild

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"devops-mcp-server/cloudbuildiface"
	"devops-mcp-server/cloudstorage"

	cloudbuild "google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// ProjectsLocationsTriggersServiceWrapper wraps cloudbuild.ProjectsLocationsTriggersService
type ProjectsLocationsTriggersServiceWrapper struct {
	*cloudbuild.ProjectsLocationsTriggersService
}

type ListResult[T any] struct {
	Items []T `json:"items"`
}
type triggersCreateCallWrapper struct {
	*cloudbuild.ProjectsLocationsTriggersCreateCall
}

func (w *triggersCreateCallWrapper) Context(ctx context.Context) cloudbuildiface.TriggersCreateCallAPI {
	w.ProjectsLocationsTriggersCreateCall.Context(ctx)
	return w
}

func (w *triggersCreateCallWrapper) Do(opts ...googleapi.CallOption) (*cloudbuild.BuildTrigger, error) {
	return w.ProjectsLocationsTriggersCreateCall.Do(opts...)
}

type triggersRunCallWrapper struct {
	*cloudbuild.ProjectsLocationsTriggersRunCall
}

func (w *triggersRunCallWrapper) Context(ctx context.Context) cloudbuildiface.TriggersRunCallAPI {
	w.ProjectsLocationsTriggersRunCall.Context(ctx)
	return w
}

func (w *triggersRunCallWrapper) Do(opts ...googleapi.CallOption) (*cloudbuild.Operation, error) {
	return w.ProjectsLocationsTriggersRunCall.Do(opts...)
}

type triggersListCallWrapper struct {
	*cloudbuild.ProjectsLocationsTriggersListCall
}

func (w *triggersListCallWrapper) Context(ctx context.Context) cloudbuildiface.TriggersListCallAPI {
	w.ProjectsLocationsTriggersListCall.Context(ctx)
	return w
}

func (w *triggersListCallWrapper) Do(opts ...googleapi.CallOption) (*cloudbuild.ListBuildTriggersResponse, error) {
	return w.ProjectsLocationsTriggersListCall.Do(opts...)
}

// Create overrides the Create method to return the correct call type.
func (w *ProjectsLocationsTriggersServiceWrapper) Create(parent string, buildtrigger *cloudbuild.BuildTrigger) cloudbuildiface.TriggersCreateCallAPI {
	return &triggersCreateCallWrapper{w.ProjectsLocationsTriggersService.Create(parent, buildtrigger)}
}

// Run overrides the Run method to return the correct call type.
func (w *ProjectsLocationsTriggersServiceWrapper) Run(name string, runbuildtriggerrequest *cloudbuild.RunBuildTriggerRequest) cloudbuildiface.TriggersRunCallAPI {
	return &triggersRunCallWrapper{w.ProjectsLocationsTriggersService.Run(name, runbuildtriggerrequest)}
}

// List overrides the List method to return the correct call type.
func (w *ProjectsLocationsTriggersServiceWrapper) List(parent string) cloudbuildiface.TriggersListCallAPI {
	return &triggersListCallWrapper{w.ProjectsLocationsTriggersService.List(parent)}
}

// ProjectsLocationsBuildsServiceWrapper wraps cloudbuild.ProjectsLocationsBuildsService
type ProjectsLocationsBuildsServiceWrapper struct {
	*cloudbuild.ProjectsLocationsBuildsService
}

type buildsCreateCallWrapper struct {
	*cloudbuild.ProjectsLocationsBuildsCreateCall
}

func (w *buildsCreateCallWrapper) Context(ctx context.Context) cloudbuildiface.BuildsCreateCallAPI {
	w.ProjectsLocationsBuildsCreateCall.Context(ctx)
	return w
}

func (w *buildsCreateCallWrapper) Do(opts ...googleapi.CallOption) (*cloudbuild.Operation, error) {
	return w.ProjectsLocationsBuildsCreateCall.Do(opts...)
}

// Create overrides the Create method to return the correct call type.
func (w *ProjectsLocationsBuildsServiceWrapper) Create(parent string, build *cloudbuild.Build) cloudbuildiface.BuildsCreateCallAPI {
	return &buildsCreateCallWrapper{w.ProjectsLocationsBuildsService.Create(parent, build)}
}

// ProjectsLocationsOperationsServiceWrapper wraps cloudbuild.ProjectsLocationsOperationsService
type ProjectsLocationsOperationsServiceWrapper struct {
	*cloudbuild.ProjectsLocationsOperationsService
}

type operationsGetCallWrapper struct {
	*cloudbuild.ProjectsLocationsOperationsGetCall
}

func (w *operationsGetCallWrapper) Context(ctx context.Context) cloudbuildiface.OperationsGetCallAPI {
	w.ProjectsLocationsOperationsGetCall.Context(ctx)
	return w
}

func (w *operationsGetCallWrapper) Do(opts ...googleapi.CallOption) (*cloudbuild.Operation, error) {
	return w.ProjectsLocationsOperationsGetCall.Do(opts...)
}

// Get overrides the Get method to return the correct call type.
func (w *ProjectsLocationsOperationsServiceWrapper) Get(name string) cloudbuildiface.OperationsGetCallAPI {
	return &operationsGetCallWrapper{w.ProjectsLocationsOperationsService.Get(name)}
}

// Client is a client for interacting with the Cloud Build API.
type Client struct {
	triggersService                  cloudbuildiface.TriggersServiceAPI
	buildsService                    cloudbuildiface.BuildsServiceAPI
	operationsService                cloudbuildiface.OperationsServiceAPI
	gcsClient                        cloudstorage.GRPClient
	regionalOperationsServiceFactory func(ctx context.Context, location string) (cloudbuildiface.OperationsServiceAPI, error)
}

// NewClient creates a new Client.
func NewClient() (*Client, error) {
	ctx := context.Background()
	service, err := cloudbuild.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud build service: %v", err)
	}

	gcsClient, err := cloudstorage.NewClient(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud storage client: %v", err)
	}

	return &Client{
		triggersService:   &ProjectsLocationsTriggersServiceWrapper{service.Projects.Locations.Triggers},
		buildsService:     &ProjectsLocationsBuildsServiceWrapper{service.Projects.Locations.Builds},
		operationsService: &ProjectsLocationsOperationsServiceWrapper{service.Projects.Locations.Operations},
		gcsClient:         gcsClient,
		regionalOperationsServiceFactory: func(ctx context.Context, location string) (cloudbuildiface.OperationsServiceAPI, error) {
			endpoint := fmt.Sprintf("%s-cloudbuild.googleapis.com", location)
			regionalService, err := cloudbuild.NewService(ctx, option.WithEndpoint(endpoint))
			if err != nil {
				return nil, fmt.Errorf("failed to create regional cloudbuild service: %w", err)
			}
			return &ProjectsLocationsOperationsServiceWrapper{regionalService.Projects.Locations.Operations}, nil
		},
	}, nil
}

// CreateTrigger creates a new Cloud Build trigger.
func (c *Client) CreateTrigger(ctx context.Context, projectID, location, triggerID, repoLink, serviceAccount, branch, tag string) (*cloudbuild.BuildTrigger, error) {
	if (branch == "") == (tag == "") {
		return nil, fmt.Errorf("exactly one of 'branch' or 'tag' must be provided")
	}

	pushConfig := &cloudbuild.PushFilter{}
	if branch != "" {
		pushConfig.Branch = branch
	}
	if tag != "" {
		pushConfig.Tag = tag
	}

	trigger := &cloudbuild.BuildTrigger{
		Name: triggerID,
		DeveloperConnectEventConfig: &cloudbuild.DeveloperConnectEventConfig{
			GitRepositoryLink: repoLink,
			Push:              pushConfig,
		},
		Autodetect:     true,
		ServiceAccount: serviceAccount,
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
	createdTrigger, err := c.triggersService.Create(parent, trigger).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create trigger: %v", err)
	}

	return createdTrigger, nil
}

// RunTrigger runs a Cloud Build trigger.
func (c *Client) RunTrigger(ctx context.Context, projectID, location, triggerID string) (*cloudbuild.Operation, error) {
	name := fmt.Sprintf("projects/%s/locations/%s/triggers/%s", projectID, location, triggerID)
	op, err := c.triggersService.Run(name, &cloudbuild.RunBuildTriggerRequest{}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to run trigger: %v", err)
	}
	return op, nil
}

// ListTriggers lists all Cloud Build triggers in a given location.
func (c *Client) ListTriggers(ctx context.Context, projectID, location string) (*ListResult[*cloudbuild.BuildTrigger], error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
	resp, err := c.triggersService.List(parent).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list triggers: %v", err)
	}
	return &ListResult[*cloudbuild.BuildTrigger]{Items: resp.Triggers}, nil
}

// BuildContainer builds a container image using Cloud Build.
func (c *Client) BuildContainer(ctx context.Context, projectID, location, repository, imageName, tag, dockerfilePath string) (*cloudbuild.Operation, error) {
	imagePath := fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s:%s", location, projectID, repository, imageName, tag)
	sourceDir := filepath.Dir(dockerfilePath)

	// Create a temporary zip file
	zipFile, err := os.CreateTemp("", "source-*.zip")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(zipFile.Name())

	// Create a new zip archive.
	writer := zip.NewWriter(zipFile)
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		zipFileWriter, err := writer.Create(relPath)
		if err != nil {
			return err
		}
		fileToZip, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fileToZip.Close()
		_, err = io.Copy(zipFileWriter, fileToZip)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk and zip source dir: %w", err)
	}
	writer.Close()
	zipFile.Close()

	// Upload the zip file to GCS
	bucketName := fmt.Sprintf("run-sources-%s-%s", projectID, location)
	objectName := fmt.Sprintf("source-%d.zip", time.Now().UnixNano())
	if err := c.gcsClient.UploadFile(ctx, projectID, bucketName, objectName, zipFile.Name()); err != nil {
		return nil, fmt.Errorf("failed to upload source to GCS: %w", err)
	}

	build := &cloudbuild.Build{
		Steps: []*cloudbuild.BuildStep{
			{
				Name: "gcr.io/cloud-builders/docker",
				Args: []string{"build", "-t", imagePath, "."},
			},
			{
				Name: "gcr.io/cloud-builders/docker",
				Args: []string{"push", imagePath},
			},
		},
		Source: &cloudbuild.Source{
			StorageSource: &cloudbuild.StorageSource{
				Bucket: bucketName,
				Object: objectName,
			},
		},
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
	op, err := c.buildsService.Create(parent, build).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create build: %v", err)
	}

	regionalOperationsService, err := c.regionalOperationsServiceFactory(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("failed to create regional cloudbuild service for operations: %w", err)
	}

	// Wait for the operation to complete
	for {
		getOp, err := regionalOperationsService.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to get operation: %v", err)
		}
		if getOp.Done {
			op = getOp
			break
		}
		time.Sleep(10 * time.Second)
	}

	if op.Error != nil {
		return nil, fmt.Errorf("build operation failed: %v", op.Error)
	}

	return op, nil
}
