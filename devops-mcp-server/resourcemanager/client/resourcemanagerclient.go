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

package resourcemanagerclient

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	resourcemanagerpb "cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
)

// contextKey is a private type to use as a key for context values.
type contextKey string

const (
	resourcemanagerClientKey contextKey = "resourcemanagerClient"
)

// ClientFrom returns the resourcemanagerClient stored in the context, if any.
func ClientFrom(ctx context.Context) (ResourcemanagerClient, bool) {
	client, ok := ctx.Value(resourcemanagerClientKey).(ResourcemanagerClient)
	return client, ok
}

// ContextWithClient returns a new context with the provided resourcemanagerClient.
func ContextWithClient(ctx context.Context, client ResourcemanagerClient) context.Context {
	return context.WithValue(ctx, resourcemanagerClientKey, client)
}

// Client is an interface for interacting with the resourcemanager API.
type ResourcemanagerClient interface {
	ToProjectNumber(ctx context.Context, projectID string) (int64, error)
}

// clientImpl is a client for interacting with the resourcemanager API.
type ResourcemanagerClientImpl struct {
	projectsClient *resourcemanager.ProjectsClient
}

// NewClient creates a new Client.
func NewClient(ctx context.Context) (ResourcemanagerClient, error) {
	rmClient, err := resourcemanager.NewProjectsClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource manager client: %v", err)
	}
	defer rmClient.Close()

	return &ResourcemanagerClientImpl{projectsClient: rmClient}, nil
}

func (r *ResourcemanagerClientImpl) ToProjectNumber(ctx context.Context, projectID string) (int64, error) {
	req := &resourcemanagerpb.GetProjectRequest{Name: fmt.Sprintf("projects/%s", projectID)}
	project, err := r.projectsClient.GetProject(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("unable to get project %s: %w", projectID, err)
	}
	projectNumber, err := strconv.ParseInt(strings.TrimPrefix(project.Name, "projects/"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse project number from %s: %w", project.Name, err)
	}
	return projectNumber, nil
}
