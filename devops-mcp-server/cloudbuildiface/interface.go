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

//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -source=interface.go -destination=mocks/mock_interface.go -package=mocks
package cloudbuildiface

import (
	"context"

	cloudbuild "google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/googleapi"
)

// BuildsCreateCallAPI defines an interface for the build creation call.
type BuildsCreateCallAPI interface {
	Context(context.Context) BuildsCreateCallAPI
	Do(opts ...googleapi.CallOption) (*cloudbuild.Operation, error)
}

// BuildsServiceAPI defines the interface for Cloud Build's Builds service.
type BuildsServiceAPI interface {
	Create(parent string, build *cloudbuild.Build) BuildsCreateCallAPI
}

// TriggersCreateCallAPI defines an interface for the trigger creation call.
type TriggersCreateCallAPI interface {
	Context(context.Context) TriggersCreateCallAPI
	Do(opts ...googleapi.CallOption) (*cloudbuild.BuildTrigger, error)
}

// TriggersRunCallAPI defines an interface for the trigger run call.
type TriggersRunCallAPI interface {
	Context(context.Context) TriggersRunCallAPI
	Do(opts ...googleapi.CallOption) (*cloudbuild.Operation, error)
}

// TriggersListCallAPI defines an interface for the trigger list call.
type TriggersListCallAPI interface {
	Context(context.Context) TriggersListCallAPI
	Do(opts ...googleapi.CallOption) (*cloudbuild.ListBuildTriggersResponse, error)
}

// TriggersServiceAPI defines the interface for Cloud Build's Triggers service.
type TriggersServiceAPI interface {
	Create(parent string, buildtrigger *cloudbuild.BuildTrigger) TriggersCreateCallAPI
	Run(name string, runbuildtriggerrequest *cloudbuild.RunBuildTriggerRequest) TriggersRunCallAPI
	List(parent string) TriggersListCallAPI
}

// OperationsGetCallAPI defines an interface for the operation get call.
type OperationsGetCallAPI interface {
	Context(context.Context) OperationsGetCallAPI
	Do(opts ...googleapi.CallOption) (*cloudbuild.Operation, error)
}

// OperationsServiceAPI defines the interface for Cloud Build's Operations service.
type OperationsServiceAPI interface {
	Get(name string) OperationsGetCallAPI
}