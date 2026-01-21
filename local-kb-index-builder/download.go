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
	"log"
	"os"
	"path/filepath"
)

// Source represents a data source to be fetched.
type Source struct {
	Name           string   `json:"name"`
	Extract        string   `json:"extract"`
	Type           string   `json:"type"`
	ExcludePattern string   `json:"exclude_pattern,omitempty"`
	Dir            string   `json:"dir,omitempty"`
	URLs           []string `json:"urls"`
	URLPattern     string   `json:"url_pattern,omitempty"`
}

var knowledgeRAGSources = []Source{
	{
		Name:           "GCP_DOCS",
		Extract:        "devsite-content",
		Type:           "webpage",
		ExcludePattern: ".*\\?hl=.+$",
		Dir:            "GCP_DOCS",
		URLs: []string{
			"https://cloud.google.com/docs/buildpacks/base-images",
			"https://cloud.google.com/docs/buildpacks/build-application",
			"https://cloud.google.com/docs/buildpacks/python",
			"https://cloud.google.com/docs/buildpacks/nodejs",
			"https://cloud.google.com/docs/buildpacks/java",
			"https://cloud.google.com/docs/buildpacks/go",
			"https://cloud.google.com/docs/buildpacks/ruby",
			"https://cloud.google.com/docs/buildpacks/php",
			"https://cloud.google.com/build/docs/build-config-file-schema",
			"https://cloud.google.com/build/docs/deploying-builds/deploy-cloud-run",
			"https://cloud.google.com/build/docs/deploying-builds/deploy-gke",
		},
	},
	{
		Name:       "Python_Specific_Docs",
		Extract:    "article",
		Type:       "webpage",
		URLPattern: ".*(?<!\\?hl=..)$",
		Dir:        "Python_Specific_Docs",
		URLs: []string{
			"https://packaging.python.org/en/latest/guides/tool-recommendations/",
			"https://packaging.python.org/en/latest/guides/section-build-and-publish/",
			"https://packaging.python.org/en/latest/tutorials/managing-dependencies/",
			"https://packaging.python.org/en/latest/tutorials/installing-packages/",
			"https://packaging.python.org/en/latest/tutorials/packaging-projects/",
			"https://packaging.python.org/en/latest/overview/",
			"https://packaging.python.org/en/latest/guides/",
			"https://packaging.python.org/en/latest/guides/tool-recommendations",
			"https://packaging.python.org/en/latest/glossary/",
			"https://packaging.python.org/en/latest/key_projects/",
			"https://py-pkgs.org/08-ci-cd.html",
			"https://switowski.com/blog/ci-101/",
		},
	},
}

func processSource(source Source, tmpDir string) {
	sourceType := source.Type

	switch sourceType {
	case "webpage":
		err := downloadWebsites(&source, tmpDir)
		if err != nil {
			log.Printf("Error downloading websites from source %s: %v", source.Name, err)
		}
	case "git_repo":
		for _, url := range source.URLs {
			repoDir := filepath.Join(tmpDir, source.Dir)
			err := fetchRepository(url, repoDir)
			if err != nil {
				log.Printf("Error downloading git repo %s: %v", url, err)
			}
		}
	default:
		log.Printf("Document Source type [%s] is not supported", sourceType)
	}
}

func processAllSources(documentSourceDir string) {
	entries, err := os.ReadDir(documentSourceDir)
	if err != nil {
		log.Fatalf("Unable to read directory: %v", err)
	}
	if len(entries) == 0 {
		for _, source := range knowledgeRAGSources {
			processSource(source, documentSourceDir)
		}
	} else {
		log.Printf("Document source directory %s is not empty, skipping download", documentSourceDir)
	}
}

func DownloadDocuments(documentSourceDir string) {

	if _, err := os.Stat(documentSourceDir); os.IsNotExist(err) {
		log.Printf("Dir does not exist: %v", documentSourceDir)
		err = os.MkdirAll(documentSourceDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Dir created: %v", documentSourceDir)
	}

	processAllSources(documentSourceDir)
}
