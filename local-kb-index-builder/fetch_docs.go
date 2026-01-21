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
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-git/go-git/v5"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

func downloadFile(url, targetDir string) (string, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	fileName := filepath.Base(url)
	filePath := filepath.Join(targetDir, fileName)

	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return filePath, err
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)

		if closeErr := outFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		if closeErr := rc.Close(); closeErr != nil && err == nil {
			err = closeErr
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func fetchRepository(repoURL, targetDir string) error {
	if strings.HasSuffix(repoURL, ".zip") {
		// It's a zip file URL
		log.Printf("Downloading zip file from %s", repoURL)

		tmpDir, err := os.MkdirTemp("", "git-zip-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDir)

		zipPath, err := downloadFile(repoURL, tmpDir)
		if err != nil {
			return fmt.Errorf("failed to download file: %w", err)
		}

		log.Printf("Unzipping %s to %s", zipPath, targetDir)
		err = unzip(zipPath, targetDir)
		if err != nil {
			return fmt.Errorf("failed to unzip file: %w", err)
		}

		log.Printf("Successfully downloaded and extracted archive from %s to %s", repoURL, targetDir)
		return nil
	}

	// It's a git repo to clone
	_, err := git.PlainClone(targetDir, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}
	log.Printf("Cloned git repository from %s to %s", repoURL, targetDir)
	return nil
}

func linkToFileName(link string, prefix string) string {
	modifiedLink := link
	if strings.HasPrefix(link, prefix) {
		modifiedLink = strings.TrimPrefix(link, prefix)
	}
	return strings.ReplaceAll(modifiedLink, "/", "_") + ".txt"
}

func convertToMarkdown(htmlContent io.Reader, element string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(htmlContent)
	if err != nil {
		return "", err
	}

	converter := md.NewConverter("", true, nil)

	var markdownContent string
	doc.Find(element).EachWithBreak(func(i int, s *goquery.Selection) bool {
		html, err := s.Html()
		if err != nil {
			log.Printf("Error getting html from selection: %v", err)
			return true // continue
		}

		markdown, err := converter.ConvertString(html)
		if err != nil {
			log.Printf("Error converting to markdown: %v", err)
			return true // continue
		}
		markdownContent = markdown
		return false // break after first element
	})

	return markdownContent, nil
}

func downloadWebsites(sources *Source, extractToDir string) error {
	urls := sources.URLs
	if len(urls) == 0 {
		return errors.New("no urls provided")
	}

	queue := make([]string, 0)
	queue = append(queue, urls...)

	fetched := make(map[string]bool)
	queued := make(map[string]bool)
	for _, u := range urls {
		queued[u] = true
	}

	extract := sources.Extract
	if extract == "" {
		return errors.New("extract field is not a string")
	}

	var excludePattern *regexp.Regexp
	if sources.ExcludePattern != "" {
		excludePattern, _ = regexp.Compile(sources.ExcludePattern)
	}

	dir := sources.Dir
	path := filepath.Join(extractToDir, dir)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}

	for len(queue) > 0 {
		currentURL := queue[0]
		queue = queue[1:]

		u, err := url.Parse(currentURL)
		if err != nil {
			log.Printf("Error parsing url %s: %v", currentURL, err)
			continue
		}
		u.Fragment = "" // remove fragment
		currentURLBase := u.String()

		if fetched[currentURLBase] {
			continue
		}

		isBaseUrl := false
		for _, baseUrl := range urls {
			if currentURLBase == baseUrl {
				isBaseUrl = true
				break
			}
		}

		if !isBaseUrl && excludePattern != nil && excludePattern.MatchString(currentURLBase) {
			log.Printf("Skipping: %s", currentURLBase)
			continue
		}

		log.Printf("Fetching: %s", currentURLBase)

		resp, err := httpClient.Get(currentURLBase)
		if err != nil {
			log.Printf("Error fetching %s: %v", currentURLBase, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Error fetching %s: %s", currentURLBase, resp.Status)
			resp.Body.Close()
			continue
		}

		fetched[currentURLBase] = true

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("Error reading body of %s: %v", currentURLBase, err)
			continue
		}

		markdownContent, err := convertToMarkdown(bytes.NewReader(bodyBytes), extract)
		if err != nil {
			log.Printf("Error converting to markdown for %s: %v", currentURLBase, err)
			continue
		}

		if markdownContent != "" {
			fileName := linkToFileName(currentURLBase, "https://")
			filePath := filepath.Join(path, fileName)
			err := os.WriteFile(filePath, []byte(markdownContent), 0644)
			if err != nil {
				log.Printf("Error writing file %s: %v", filePath, err)
			}
		}

		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bodyBytes))
		if err != nil {
			log.Printf("Error parsing html from %s: %v", currentURLBase, err)
			continue
		}

		baseURL, err := url.Parse(currentURLBase)
		if err != nil {
			log.Printf("Error parsing base url %s: %v", currentURLBase, err)
			continue
		}

		doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
			link, exists := s.Attr("href")
			if !exists {
				return
			}

			absoluteLink, err := baseURL.Parse(link)
			if err != nil {
				return
			}
			absoluteLink.Fragment = ""
			absoluteLinkBase := absoluteLink.String()

			isInternal := false
			for _, u := range urls {
				if strings.HasPrefix(absoluteLinkBase, u) {
					isInternal = true
					break
				}
			}

			if isInternal {
				if !queued[absoluteLinkBase] {
					queued[absoluteLinkBase] = true
					queue = append(queue, absoluteLinkBase)
				}
			}
		})
	}
	return nil
}
