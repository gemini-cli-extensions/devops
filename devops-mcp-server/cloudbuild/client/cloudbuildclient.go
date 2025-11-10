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

package cloudbuildclient

import (
	"context"
	"fmt"

	cloudbuild "cloud.google.com/go/cloudbuild/apiv1/v2"
	cloudbuildpb "cloud.google.com/go/cloudbuild/apiv1/v2/cloudbuildpb"

	build "google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// contextKey is a private type to use as a key for context values.
type contextKey string

const (
	cloudBuildClientContextKey contextKey = "cloudBuildClient"
)

// ClientFrom returns the CloudBuildClient stored in the context, if any.
func ClientFrom(ctx context.Context) (CloudBuildClient, bool) {
	client, ok := ctx.Value(cloudBuildClientContextKey).(CloudBuildClient)
	return client, ok
}

// ContextWithClient returns a new context with the provided CloudBuildClient.
func ContextWithClient(ctx context.Context, client CloudBuildClient) context.Context {
	return context.WithValue(ctx, cloudBuildClientContextKey, client)
}

// CloudBuildClient is an interface for interacting with the Cloud Build API.
type CloudBuildClient interface {
	CreateBuildTrigger(ctx context.Context, projectID, location, triggerID, repoLink, branch, tag, serviceAccount string) (*build.BuildTrigger, error)
	GetLatestBuildForTrigger(ctx context.Context, projectID, location, triggerID string) (*cloudbuildpb.Build, error)
	ListBuildTriggers(ctx context.Context, projectID, location string) ([]*cloudbuildpb.BuildTrigger, error)
	RunBuildTrigger(ctx context.Context, projectID, location, triggerID, branch, tag, commitSha string) (*cloudbuild.RunBuildTriggerOperation, error)
}

// NewCloudBuildClient creates a new Cloud Build client.
func NewCloudBuildClient(ctx context.Context) (CloudBuildClient, error) {
	c, err := cloudbuild.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Build client: %v", err)
	}

	c2, err := build.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Build service: %v", err)
	}

	return &CloudBuildClientImpl{v1client: c, legacyClient: c2}, nil
}

// CloudBuildClientImpl is an implementation of the CloudBuildClient interface.
type CloudBuildClientImpl struct {
	v1client     *cloudbuild.Client
	legacyClient *build.Service
}

// CreateCloudBuildTrigger creates a new build trigger.
func (c *CloudBuildClientImpl) CreateBuildTrigger(ctx context.Context, projectID, location, triggerID, repoLink, branch, tag, serviceAccount string) (*build.BuildTrigger, error) {
	if (branch == "") == (tag == "") {
		return nil, fmt.Errorf("exactly one of 'branch' or 'tag' must be provided")
	}

	trigger := &build.BuildTrigger{
		Name: triggerID,
		TriggerTemplate: &build.RepoSource{
			RepoName: repoLink, // This should be the Developer Connect repo link
		},
		ServiceAccount: serviceAccount,
		Autodetect:     true,
	}

	if branch != "" {
		trigger.TriggerTemplate.BranchName = branch
	} else {
		trigger.TriggerTemplate.TagName = tag
	}

	return c.legacyClient.Projects.Locations.Triggers.Create(fmt.Sprintf("projects/%s/locations/%s", projectID, location), trigger).Context(ctx).Do()
}

func (c *CloudBuildClientImpl) GetLatestBuildForTrigger(ctx context.Context, projectID, location, triggerID string) (*cloudbuildpb.Build, error) {
	req := &cloudbuildpb.ListBuildsRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", projectID, location),
		Filter: fmt.Sprintf("trigger_id = %q", triggerID),
	}
	it := c.v1client.ListBuilds(ctx, req) // Uses v1client
	var latestBuild *cloudbuildpb.Build
	var latestTime *timestamppb.Timestamp

	for {
		build, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list builds: %v", err)
		}

		if latestTime == nil || build.CreateTime.AsTime().After(latestTime.AsTime()) {
			latestTime = build.CreateTime
			latestBuild = build
		}
	}

	if latestBuild == nil {
		return nil, fmt.Errorf("no builds found for trigger ID: %s", triggerID)
	}

	return latestBuild, nil
}

// ListBuildTriggers lists all build triggers for a given project.
func (c *CloudBuildClientImpl) ListBuildTriggers(ctx context.Context, projectID, location string) ([]*cloudbuildpb.BuildTrigger, error) {
	req := &cloudbuildpb.ListBuildTriggersRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", projectID, location),
	}
	it := c.v1client.ListBuildTriggers(ctx, req)
	var triggers []*cloudbuildpb.BuildTrigger
	for {
		trigger, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list build triggers: %v", err)
		}
		triggers = append(triggers, trigger)
	}
	return triggers, nil
}

// RunBuildTrigger runs a build trigger.
func (c *CloudBuildClientImpl) RunBuildTrigger(ctx context.Context, projectID, location, triggerID, branch, tag, commitSha string) (*cloudbuild.RunBuildTriggerOperation, error) {
	if (branch == "") == (tag == "") == (commitSha == "") {
		return nil, fmt.Errorf("exactly one of 'branch' or 'tag' or 'commitSha' must be provided")
	}
	req := &cloudbuildpb.RunBuildTriggerRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/triggers/%s", projectID, location, triggerID),
	}
	if branch != "" {
		req.Source = &cloudbuildpb.RepoSource{
			Revision: &cloudbuildpb.RepoSource_BranchName{BranchName: branch},
		}
	} else if tag != "" {
		req.Source = &cloudbuildpb.RepoSource{
			Revision: &cloudbuildpb.RepoSource_TagName{TagName: tag},
		}
	} else {
		req.Source = &cloudbuildpb.RepoSource{
			Revision: &cloudbuildpb.RepoSource_CommitSha{CommitSha: commitSha},
		}
	}
	op, err := c.v1client.RunBuildTrigger(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to run build trigger: %v", err)
	}
	return op, nil
}
