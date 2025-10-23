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

package containeranalysis

import (
	"context"
	"fmt"

	containeranalysis "cloud.google.com/go/containeranalysis/apiv1"
	"google.golang.org/api/iterator"
	grafeaspb "google.golang.org/genproto/googleapis/grafeas/v1"
)



// ListResult defines a generic struct to wrap a list of items.

type ListResult[T any] struct {

	Items []T `json:"items"`

}



// Client is a client for interacting with the Container Analysis API.

type Client struct {

	client *containeranalysis.Client

}



// NewClient creates a new Client.

func NewClient(ctx context.Context) (*Client, error) {

	c, err := containeranalysis.NewClient(ctx)

	if err != nil {

		return nil, fmt.Errorf("failed to create container analysis client: %v", err)

	}

	return &Client{client: c}, nil

}



// ListVulnerabilities lists vulnerabilities for a given image resource URL.

func (c *Client) ListVulnerabilities(ctx context.Context, projectID, resourceURL string) (*ListResult[*grafeaspb.Occurrence], error) {

	req := &grafeaspb.ListOccurrencesRequest{

		Parent: fmt.Sprintf("projects/%s", projectID),

		Filter: fmt.Sprintf("resourceUrl=\"%s\" AND kind=\"VULNERABILITY\"", resourceURL),

	}

	it := c.client.GetGrafeasClient().ListOccurrences(ctx, req)

	var vulnerabilities []*grafeaspb.Occurrence

	for {

		occurrence, err := it.Next()

		if err == iterator.Done {

			break

		}

		if err != nil {

			return nil, fmt.Errorf("failed to list vulnerabilities: %v", err)

		}

		vulnerabilities = append(vulnerabilities, occurrence)

	}

	return &ListResult[*grafeaspb.Occurrence]{Items: vulnerabilities}, nil

}
