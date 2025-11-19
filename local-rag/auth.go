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

package main

import (
	"context"
	"fmt"

	"cloud.google.com/go/auth/credentials"
)

// getGCPToken retrieves the Google Cloud Platform access token and project ID
// using Application Default Credentials.
func getGCPToken(ctx context.Context) (tokenValue string, projectID string, err error) {
	// Use Application Default Credentials to get a TokenSource
	scopes := []string{"https://www.googleapis.com/auth/cloud-platform"}
	creds, err := credentials.DetectDefault(&credentials.DetectOptions{
		Scopes: scopes,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to find default credentials: %w", err)
	}

	projectID, err = creds.ProjectID(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get project ID: %w", err)
	}

	if projectID == "" {
		// Try quota project
		projectID, err = creds.QuotaProjectID(ctx)
		if err != nil {
			return "", "", fmt.Errorf("failed to get quota project ID: %w", err)
		}
		if projectID == "" {
			return "", "", fmt.Errorf("no Project ID found in Application Default Credentials. " +
				"This can happen if credentials are user-based or the project hasn't been explicitly set " +
				"e.g., via gcloud auth application-default set-quota-project")
		}
	}

	// We need an access token
	token, err := creds.TokenProvider.Token(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to retrieve access token: %w", err)
	}

	return token.Value, projectID, nil
}
