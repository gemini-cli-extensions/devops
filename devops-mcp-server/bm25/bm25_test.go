// Copyright 2025 Google LLC
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

package bm25

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"devops-mcp-server/bm25/client/mocks"

	"github.com/golang/mock/gomock"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBM25Client := mocks.NewMockBM25Client(ctrl)

	handler := &Handler{
		BM25Client: mockBM25Client,
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "test"}, &mcp.ServerOptions{})
	handler.Register(server)

	// Verify that the tools were added to the server
	if queryPatternToolFunc == nil {
		t.Error("queryPatternToolFunc was not initialized")
	}
	if queryKnowledgeToolFunc == nil {
		t.Error("queryKnowledgeToolFunc was not initialized")
	}
}

func TestQueryPatternTool(t *testing.T) {
	ctx := context.Background()
	query := "test query"

	tests := []struct {
		name           string
		setupMocks     func(*mocks.MockBM25Client)
		expectErr      bool
		expectedResult any
		expectedError  string
	}{
		{
			name: "Success case",
			setupMocks: func(mock *mocks.MockBM25Client) {
				mock.EXPECT().QueryPatterns(gomock.Any(), query).Return("mocked pattern result", nil)
			},
			expectErr:      false,
			expectedResult: map[string]any{"cicd-patterns": "mocked pattern result"},
		},
		{
			name: "Error case",
			setupMocks: func(mock *mocks.MockBM25Client) {
				mock.EXPECT().QueryPatterns(gomock.Any(), query).Return("", errors.New("query failed"))
			},
			expectErr:     true,
			expectedError: "failed to query patterns: query failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockBM25Client := mocks.NewMockBM25Client(ctrl)
			tt.setupMocks(mockBM25Client)

			server := mcp.NewServer(&mcp.Implementation{Name: "test"}, &mcp.ServerOptions{})
			addQueryPatternTool(server, mockBM25Client)

			_, res, err := queryPatternToolFunc(ctx, nil, QueryArgs{Query: query})

			if (err != nil) != tt.expectErr {
				t.Errorf("queryPatternToolFunc() error = %v, expectErr %v", err, tt.expectErr)
			}

			if tt.expectErr && err.Error() != tt.expectedError {
				t.Errorf("queryPatternToolFunc() error = %q, expectedError %q", err.Error(), tt.expectedError)
			}

			if !tt.expectErr && !reflect.DeepEqual(res, tt.expectedResult) {
				t.Errorf("queryPatternToolFunc() result = %v, expectedResult %v", res, tt.expectedResult)
			}
		})
	}
}

func TestQueryKnowledgeTool(t *testing.T) {
	ctx := context.Background()
	query := "test query"

	tests := []struct {
		name           string
		setupMocks     func(*mocks.MockBM25Client)
		expectErr      bool
		expectedResult any
		expectedError  string
	}{
		{
			name: "Success case",
			setupMocks: func(mock *mocks.MockBM25Client) {
				mock.EXPECT().Queryknowledge(gomock.Any(), query).Return("mocked knowledge result", nil)
			},
			expectErr:      false,
			expectedResult: map[string]any{"knowledge": "mocked knowledge result"},
		},
		{
			name: "Error case",
			setupMocks: func(mock *mocks.MockBM25Client) {
				mock.EXPECT().Queryknowledge(gomock.Any(), query).Return("", errors.New("query failed"))
			},
			expectErr:     true,
			expectedError: "failed to query knowledge: query failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockBM25Client := mocks.NewMockBM25Client(ctrl)
			tt.setupMocks(mockBM25Client)

			server := mcp.NewServer(&mcp.Implementation{Name: "test"}, &mcp.ServerOptions{})
			addQueryKnowledgeTool(server, mockBM25Client)

			_, res, err := queryKnowledgeToolFunc(ctx, nil, QueryArgs{Query: query})

			if (err != nil) != tt.expectErr {
				t.Errorf("queryKnowledgeToolFunc() error = %v, expectErr %v", err, tt.expectErr)
			}

			if tt.expectErr && err.Error() != tt.expectedError {
				t.Errorf("queryKnowledgeToolFunc() error = %q, expectedError %q", err.Error(), tt.expectedError)
			}

			if !tt.expectErr && !reflect.DeepEqual(res, tt.expectedResult) {
				t.Errorf("queryKnowledgeToolFunc() result = %v, expectedResult %v", res, tt.expectedResult)
			}
		})
	}
}
