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

	cloudbuild "google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/googleapi"
)

// GRPCClient is an interface for interacting with the Cloud Build API.
type GRPClient interface {
	CreateTrigger(ctx context.Context, projectID, location, triggerID, repoLink, serviceAccount, branch, tag string) (*cloudbuild.BuildTrigger, error)
	RunTrigger(ctx context.Context, projectID, location, triggerID string) (*cloudbuild.Operation, error)
	ListTriggers(ctx context.Context, projectID, location string) ([]*cloudbuild.BuildTrigger, error)
	BuildContainer(ctx context.Context, projectID, location, repository, imageName, tag, dockerfilePath string) (*cloudbuild.Operation, error)
}

// TriggersServiceAPI defines the interface for the Cloud Build Triggers service.
type TriggersServiceAPI interface {
	Create(parent string, buildtrigger *cloudbuild.BuildTrigger) TriggersCreateCallAPI
	Run(name string, runbuildtriggerrequest *cloudbuild.RunBuildTriggerRequest) TriggersRunCallAPI
	List(parent string) TriggersListCallAPI
}

// TriggersCreateCallAPI defines the interface for a triggers create call.
type TriggersCreateCallAPI interface {
	Context(context.Context) TriggersCreateCallAPI
	Do(...googleapi.CallOption) (*cloudbuild.BuildTrigger, error)
}

// TriggersRunCallAPI defines the interface for a triggers run call.
type TriggersRunCallAPI interface {
	Context(context.Context) TriggersRunCallAPI
	Do(...googleapi.CallOption) (*cloudbuild.Operation, error)
}

// TriggersListCallAPI defines the interface for a triggers list call.
type TriggersListCallAPI interface {
	Context(context.Context) TriggersListCallAPI
	Do(...googleapi.CallOption) (*cloudbuild.ListBuildTriggersResponse, error)
}

// BuildsServiceAPI defines the interface for the Cloud Build Builds service.
type BuildsServiceAPI interface {
	Create(parent string, build *cloudbuild.Build) BuildsCreateCallAPI
}

// BuildsCreateCallAPI defines the interface for a builds create call.
type BuildsCreateCallAPI interface {
	Context(context.Context) BuildsCreateCallAPI
	Do(...googleapi.CallOption) (*cloudbuild.Operation, error)
}
