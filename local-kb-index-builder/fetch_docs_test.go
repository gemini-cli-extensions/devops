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
	"archive/zip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDownloadFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("test content"))
	}))
	defer server.Close()

	tmpDir, err := os.MkdirTemp("", "test-download-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath, err := downloadFile(server.URL, tmpDir)
	if err != nil {
		t.Fatalf("downloadFile failed: %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != "test content" {
		t.Errorf("Expected 'test content', got '%s'", string(content))
	}
}

func TestUnzip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-unzip-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	zipPath := filepath.Join(tmpDir, "test.zip")
	destPath := filepath.Join(tmpDir, "unzipped")

	// Create a dummy zip file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	zipWriter := zip.NewWriter(zipFile)
	writer, err := zipWriter.Create("test.txt")
	if err != nil {
		t.Fatalf("Failed to create entry in zip file: %v", err)
	}
	_, err = writer.Write([]byte("test content"))
	if err != nil {
		t.Fatalf("Failed to write to entry in zip file: %v", err)
	}
	zipWriter.Close()
	zipFile.Close()

	err = unzip(zipPath, destPath)
	if err != nil {
		t.Fatalf("unzip failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destPath, "test.txt"))
	if err != nil {
		t.Fatalf("Failed to read unzipped file: %v", err)
	}

	if string(content) != "test content" {
		t.Errorf("Expected 'test content', got '%s'", string(content))
	}
}

func TestDownloadWebsites(t *testing.T) {
	// Mock server to simulate website structure
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_, _ = w.Write([]byte(`<html><body><devsite-content><a href="/page2">Page 2</a></devsite-content></body></html>`))
		} else if r.URL.Path == "/page2" {
			_, _ = w.Write([]byte(`<html><body><devsite-content>Content of Page 2</devsite-content></body></html>`))
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tmpDir, err := os.MkdirTemp("", "test-websites-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sources := &Source{
		URLs:    []string{server.URL},
		Extract: "devsite-content",
		Dir:     "test_docs",
	}

	err = downloadWebsites(sources, tmpDir)
	if err != nil {
		t.Fatalf("downloadWebsites failed: %v", err)
	}

	// Check if files were created
	files, err := os.ReadDir(filepath.Join(tmpDir, "test_docs"))
	if err != nil {
		t.Fatalf("Failed to read output directory: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}

	// Check content of one of the files
	// The filename is based on the URL, which is unpredictable with httptest.
	// So we just check if one of the files has the expected content.
	foundPage2 := false
	for _, file := range files {
		if !file.IsDir() {
			content, err := os.ReadFile(filepath.Join(tmpDir, "test_docs", file.Name()))
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}
			if strings.Contains(string(content), "Content of Page 2") {
				foundPage2 = true
			}
		}
	}
	if !foundPage2 {
		t.Error("Did not find the content of page 2 in any of the downloaded files")
	}
}

func TestConvertToMarkdown(t *testing.T) {
	html := `<html><body><div id="content"><h1>Title</h1><p>Some text.</p></div></body></html>`
	reader := strings.NewReader(html)

	markdown, err := convertToMarkdown(reader, "#content")
	if err != nil {
		t.Fatalf("convertToMarkdown failed: %v", err)
	}

	expected := "# Title\n\nSome text."
	// The converter might add extra newlines, so we trim space
	if strings.TrimSpace(markdown) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, strings.TrimSpace(markdown))
	}
}

func TestLinkToFileName(t *testing.T) {
	link := "https://example.com/path/to/file"
	prefix := "https://"
	expected := "example.com_path_to_file.txt"
	result := linkToFileName(link, prefix)
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// A helper function to create a dummy zip file for testing downloadGitRepo with a zip URL
func createDummyZip(t *testing.T, path string) {
	t.Helper()
	zipFile, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer zipFile.Close()
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	writer, err := zipWriter.Create("dummy.txt")
	if err != nil {
		t.Fatalf("Failed to create entry in zip file: %v", err)
	}
	_, err = writer.Write([]byte("dummy content"))
	if err != nil {
		t.Fatalf("Failed to write to entry in zip file: %v", err)
	}
}

func TestFetchRepository_Zip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-zip-server-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dummyZipPath := filepath.Join(tmpDir, "test.zip")
	createDummyZip(t, dummyZipPath)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, dummyZipPath)
	}))
	defer server.Close()

	targetDir := filepath.Join(tmpDir, "target")
	err = fetchRepository(server.URL+"/test.zip", targetDir)
	if err != nil {
		t.Fatalf("fetchRepository with zip failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(targetDir, "dummy.txt"))
	if err != nil {
		t.Fatalf("Failed to read file from unzipped archive: %v", err)
	}
	if string(content) != "dummy content" {
		t.Errorf("Expected 'dummy content', got '%s'", string(content))
	}
}
