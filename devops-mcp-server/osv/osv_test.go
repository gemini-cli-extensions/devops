
// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package osv

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	
	osvmocks "devops-mcp-server/osv/client/mocks"
)

func TestAddScanSecretsTool(t *testing.T) {
	ctx := context.Background()
	root := "/test/dir"

	tests := []struct {
		name          string
		args          ScanSecretsArgs
		setupMocks    func(*osvmocks.MockOsvClient)
		expectErr     bool
		expectedError string
		expectedResult string
	}{
		{
			name: "Success case",
			args: ScanSecretsArgs{
				Root: root,
			},
			setupMocks: func(osvMock *osvmocks.MockOsvClient) {
				osvMock.EXPECT().ScanSecrets(gomock.Any(), root).Return("scan results", nil)
			},
			expectErr: false,
			expectedResult: "scan results",
		},
		{
			name: "Error case",
			args: ScanSecretsArgs{
				Root: root,
			},
			setupMocks: func(osvMock *osvmocks.MockOsvClient) {
				osvMock.EXPECT().ScanSecrets(gomock.Any(), root).Return("", errors.New("scan failed"))
			},
			expectErr:     true,
			expectedError: "failed to scan for secrets: scan failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			osvMock := osvmocks.NewMockOsvClient(ctrl)
			tt.setupMocks(osvMock)

			server := mcp.NewServer(&mcp.Implementation{Name: "test"}, &mcp.ServerOptions{})
			addScanSecretsTool(server, osvMock)

			_, res, err := scanSecretsToolFunc(ctx, nil, tt.args)

			if (err != nil) != tt.expectErr {
				t.Errorf("scanSecretsToolFunc() error = %v, expectErr %v", err, tt.expectErr)
			}

			if tt.expectErr && err.Error() != tt.expectedError {
				t.Errorf("scanSecretsToolFunc() error = %q, expectedError %q", err.Error(), tt.expectedError)
			}
			
			if !tt.expectErr {
				resultMap, ok := res.(map[string]any)
				if !ok {
					t.Fatalf("Unexpected result type: %T", res)
				}
				report, ok := resultMap["report"].(string)
				if !ok {
					t.Fatalf("Unexpected report type: %T", resultMap["report"])
				}
				if report != tt.expectedResult {
					t.Errorf("scanSecretsToolFunc() result = %q, expectedResult %q", report, tt.expectedResult)
				}
			}
		})
	}
}


func TestHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	osvMock := osvmocks.NewMockOsvClient(ctrl)

	handler := &Handler{
		OsvClient: osvMock,
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "test"}, &mcp.ServerOptions{})
	handler.Register(server)

	// Verify that the tool was added to the server
	if scanSecretsToolFunc == nil {
		t.Error("scanSecretsToolFunc was not initialized")
	}
}
