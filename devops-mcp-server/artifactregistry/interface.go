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

package artifactregistry

import (
	"context"

	artifactregistrypb "google.golang.org/genproto/googleapis/devtools/artifactregistry/v1"
)

// CreateRepositoryOperation is an interface that wraps the artifactregistry.CreateRepositoryOperation.
type CreateRepositoryOperation interface {
	Wait(context.Context) (*artifactregistrypb.Repository, error)
}

// GRPClient is an interface that wraps the artifactregistry.Client.
type GRPClient interface {
	GetRepository(context.Context, *artifactregistrypb.GetRepositoryRequest) (*artifactregistrypb.Repository, error)
	CreateRepository(context.Context, *artifactregistrypb.CreateRepositoryRequest) (CreateRepositoryOperation, error)
}
