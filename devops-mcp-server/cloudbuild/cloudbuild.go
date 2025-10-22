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
	"context"
	"fmt"

	"google.golang.org/api/cloudbuild/v1"
)

// ListResult defines a generic struct to wrap a list of items.
type ListResult[T any] struct {
	Items []T `json:"items"`
}

// Client is a client for interacting with the Cloud Build API.
type Client struct {
	service *cloudbuild.Service
}

// NewClient creates a new Client.
func NewClient() (*Client, error) {
	ctx := context.Background()
	service, err := cloudbuild.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud build service: %v", err)
	}
	return &Client{service: service}, nil
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
	createdTrigger, err := c.service.Projects.Locations.Triggers.Create(parent, trigger).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create trigger: %v", err)
	}

	return createdTrigger, nil
}

// RunTrigger runs a Cloud Build trigger.
func (c *Client) RunTrigger(ctx context.Context, projectID, location, triggerID string) (*cloudbuild.Operation, error) {
	name := fmt.Sprintf("projects/%s/locations/%s/triggers/%s", projectID, location, triggerID)
	op, err := c.service.Projects.Locations.Triggers.Run(name, &cloudbuild.RunBuildTriggerRequest{}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to run trigger: %v", err)
	}
	return op, nil
}

// ListTriggers lists all Cloud Build triggers in a given location.
func (c *Client) ListTriggers(ctx context.Context, projectID, location string) (*ListResult[*cloudbuildv1.BuildTrigger], error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
	resp, err := c.service.Projects.Locations.Triggers.List(parent).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list triggers: %v", err)
	}
	return &ListResult[*cloudbuildv1.BuildTrigger]{Items: resp.Triggers}, nil
}
