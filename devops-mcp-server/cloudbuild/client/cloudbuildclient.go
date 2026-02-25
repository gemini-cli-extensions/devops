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
	"strings"

	cloudbuild "cloud.google.com/go/cloudbuild/apiv1/v2"
	logging "cloud.google.com/go/logging/apiv2"
	build "google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	
	cloudbuildpb "cloud.google.com/go/cloudbuild/apiv1/v2/cloudbuildpb"
	loggingpb "cloud.google.com/go/logging/apiv2/loggingpb"
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
	ListBuilds(ctx context.Context, projectID, location string) ([]*cloudbuildpb.Build, error)
	GetBuildInfo(ctx context.Context, projectID, location, buildID string) (BuildInfo, error)
	StartBuild(ctx context.Context, projectID, location string, source *cloudbuildpb.Source) (*cloudbuild.CreateBuildOperation, error)
}

type BuildInfo struct {
	BuildDetails *cloudbuildpb.Build
	Logs string
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

	loggingClient, err := logging.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Logging client: %v", err)
	}

	return &CloudBuildClientImpl{
		v1client:      c,
		legacyClient:  c2,
		loggingClient: loggingClient,
	}, nil
}

// CloudBuildClientImpl is an implementation of the CloudBuildClient interface.
type CloudBuildClientImpl struct {
	v1client     *cloudbuild.Client
	legacyClient *build.Service
	loggingClient *logging.Client
}

// CreateCloudBuildTrigger creates a new build trigger.
func (c *CloudBuildClientImpl) CreateBuildTrigger(ctx context.Context, projectID, location, triggerID, repoLink, branch, tag, serviceAccount string) (*build.BuildTrigger, error) {
	if (branch == "") == (tag == "") {
		return nil, fmt.Errorf("exactly one of 'branch' or 'tag' must be provided")
	}

	pushFilter := &build.PushFilter{}
	if branch != "" {
		pushFilter.Branch = branch

	} else {
		pushFilter.Tag = tag
	}
	sa := strings.TrimPrefix(serviceAccount, "serviceAccount:")
	sa = fmt.Sprintf("projects/%s/serviceAccounts/%s", projectID, sa)
	trigger := &build.BuildTrigger{
		Name: triggerID,
		DeveloperConnectEventConfig: &build.DeveloperConnectEventConfig{
			GitRepositoryLink:     repoLink, // This should be the Developer Connect repo link
			Push:                  pushFilter,
			GitRepositoryLinkType: "GITHUB",
		},
		ServiceAccount: sa,
		Filename:       "cloudbuild.yaml",
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


func (c *CloudBuildClientImpl) ListBuilds(ctx context.Context, projectID, location string) ([]*cloudbuildpb.Build, error) {
	req := &cloudbuildpb.ListBuildsRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", projectID, location),
	}
	it := c.v1client.ListBuilds(ctx, req)
	var builds []*cloudbuildpb.Build
	for {
		build, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
return nil, fmt.Errorf("failed to list builds: %w", err)
		}
		builds = append(builds, build)
	}
	return builds, nil
}

func (c *CloudBuildClientImpl) GetBuildInfo(ctx context.Context, projectID, location, buildID string) (BuildInfo, error) {
	req := &cloudbuildpb.GetBuildRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/builds/%s", projectID, location, buildID),
	}
	build, err := c.v1client.GetBuild(ctx, req)
	if err != nil {
		return BuildInfo{}, fmt.Errorf("failed to get build info: %w", err)
	}
	info := BuildInfo{BuildDetails: build}
	logReq := &loggingpb.ListLogEntriesRequest{
		ResourceNames: []string{fmt.Sprintf("projects/%s", projectID)},
		Filter:        fmt.Sprintf(`resource.type="build" AND resource.labels.build_id="%s" AND logName="projects/%s/logs/cloudbuild"`, buildID, projectID),
	}
	it := c.loggingClient.ListLogEntries(ctx, logReq)
	var logs []string
	for {
		entry, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return BuildInfo{}, fmt.Errorf("failed to list log entries: %w", err)
		}
		var logMessage string
		switch payload := entry.Payload.(type) {
			case *loggingpb.LogEntry_TextPayload:
				logMessage = payload.TextPayload
			case *loggingpb.LogEntry_JsonPayload:
				jsonBytes, err := protojson.Marshal(payload.JsonPayload)
				if err != nil {
					logMessage = fmt.Sprintf("failed to marshal json payload to string: %v", err)
				} else {
					logMessage = string(jsonBytes)
				}
			case *loggingpb.LogEntry_ProtoPayload:
				logMessage = fmt.Sprintf("%v", payload.ProtoPayload)
			default:
				return BuildInfo{}, fmt.Errorf("unknown log entry payload type")
			}
		logs = append(logs, logMessage)
	}
	info.Logs = strings.Join(logs, "\n")
	return info, nil
}

func (c *CloudBuildClientImpl) StartBuild(ctx context.Context, projectID, location string, source *cloudbuildpb.Source) (*cloudbuild.CreateBuildOperation, error) {
	req := &cloudbuildpb.CreateBuildRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", projectID, location),
		Build:  &cloudbuildpb.Build{Source: source},
	}
	ops, err := c.v1client.CreateBuild(ctx, req)
	if err != nil {
return nil, fmt.Errorf("failed to start build: %w", err)
	}
	return ops, nil
}
