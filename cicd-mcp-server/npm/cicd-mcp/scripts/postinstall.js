// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

'use strict';

const fs = require('fs');
const path = require('path');
const os = require('os');

const PLATFORM = os.platform();
const ARCH = os.arch();

// 1. Use a lookup table instead of if-else chains
const SUPPORTED_PACKAGES = {
  'linux_x64': '@google-cloud/cicd-mcp-linux-amd64',
  'darwin_arm64': '@google-cloud/cicd-mcp-darwin-arm64',
  'win32_x64': '@google-cloud/cicd-mcp-windows-amd64',
};

const key = `${PLATFORM}_${ARCH}`;
const subPackage = SUPPORTED_PACKAGES[key];

if (!subPackage) {
  console.error(`Unsupported platform/architecture: ${PLATFORM}/${ARCH}`);
  process.exit(1);
}

// 2. Derive paths using const
const isWindows = PLATFORM === 'win32';
const binaryName = isWindows ? 'cicd-mcp-server.exe' : 'cicd-mcp-server';

// In a typical install, the sub-package will be in the same node_modules/@google-cloud directory
// __dirname is node_modules/@google-cloud/cicd-mcp/scripts
// We go up 3 levels to reach node_modules
const sourcePath = path.join(__dirname, '..', '..', '..', subPackage, 'bin', binaryName);
const targetPath = path.join(__dirname, '..', 'bin', binaryName);

try {
  // Check if source exists
  if (!fs.existsSync(sourcePath)) {
    console.error(`Source binary not found at ${sourcePath}`);
    process.exit(1);
  }

  // Remove placeholder if it exists
  if (fs.existsSync(targetPath)) {
    fs.unlinkSync(targetPath);
  }

  // Create hard link or copy
  try {
    fs.linkSync(sourcePath, targetPath);
    console.log(`Linked ${sourcePath} to ${targetPath}`);
  } catch (err) {
    // If linking fails (e.g. cross-device), copy instead
    fs.copyFileSync(sourcePath, targetPath);
    console.log(`Copied ${sourcePath} to ${targetPath}`);
  }

  // Ensure executable permissions (not needed on Windows)
  if (!isWindows) {
    fs.chmodSync(targetPath, 0o755);
  }
} catch (err) {
  console.error(`Failed to install binary: ${err.message}`);
  process.exit(1);
}
