// plugin.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Tag struct {
	Name string `json:"name"`
}

type Content struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"` // "file" o "dir"
	DownloadURL string `json:"download_url"` // solo per file
}

var versionRe = regexp.MustCompile(`^v(\d+\.\d+\.\d+)$`)

// Scarica tutti i .proto da tutte le versioni >= v2.0.14 e compila in .pb.go
func DownloadAllProtos(githubToken string) error {
	owner := "meshtastic"
	repo := "protobufs"
	basePath := "meshtastic"

	tags, err := listTags(owner, repo, githubToken)
	if err != nil {
		return fmt.Errorf("listTags: %w", err)
	}
	filtered := filterTags(tags, "v2.0.14")
	fmt.Printf("Trovati %d tag >= v2.0.14\n", len(filtered))
	for _, tag := range filtered {
		fmt.Printf("-> Scarico proto per tag %s\n", tag)
		localDir := filepath.Join("pb", "meshtastic-"+tag)
		if err := downloadDir(owner, repo, basePath, tag, localDir, githubToken); err != nil {
			return fmt.Errorf("errore downloadDir per %s: %w", tag, err)
		}
		// Compila tutti i .proto in questa directory in .pb.go (per tutti i file presenti)
		if err := compileProtos(localDir); err != nil {
			return fmt.Errorf("errore compileProtos %s: %w", tag, err)
		}
	}
	return nil
}

func listTags(owner, repo, githubToken string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/tags?per_page=100", owner, repo)
	req, _ := http.NewRequest("GET", url, nil)
	if githubToken != "" {
		req.Header.Set("Authorization", "token "+githubToken)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var tags []Tag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}
	var names []string
	for _, t := range tags {
		names = append(names, t.Name)
	}
	sort.Strings(names)
	return names, nil
}

func filterTags(tags []string, min string) []string {
	var out []string
	for _, t := range tags {
		if versionRe.MatchString(t) && compareVer(t, min) >= 0 {
			out = append(out, t)
		}
	}
	return out
}

// compareVer("v2.0.15", "v2.0.14") > 0
func compareVer(a, b string) int {
	av := strings.TrimPrefix(a, "v")
	bv := strings.TrimPrefix(b, "v")
	return strings.Compare(av, bv)
}

func downloadDir(owner, repo, dirPath, ref, localDir, githubToken string) error {
	apiURL := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
		owner, repo, dirPath, ref,
	)
	req, _ := http.NewRequest("GET", apiURL, nil)
	if githubToken != "" {
		req.Header.Set("Authorization", "token "+githubToken)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
	}

	var items []Content
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return err
	}

	// Assicura che la cartella locale esista
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return err
	}

	for _, item := range items {
		localPath := filepath.Join(localDir, item.Name)
		switch item.Type {
		case "file":
			if strings.HasSuffix(item.Name, ".proto") {
				if err := downloadFile(item.DownloadURL, localPath, githubToken); err != nil {
					return err
				}
				fmt.Printf("✔ Scaricato %s\n", item.Path)
			}
		case "dir":
			if err := downloadDir(owner, repo, item.Path, ref, filepath.Join(localDir, item.Name), githubToken); err != nil {
				return err
			}
		default:
			fmt.Printf("⚠ Ignoro %s di tipo %s\n", item.Path, item.Type)
		}
	}
	return nil
}

func downloadFile(url, dest, githubToken string) error {
	req, _ := http.NewRequest("GET", url, nil)
	if githubToken != "" {
		req.Header.Set("Authorization", "token "+githubToken)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download %s -> status %d: %s", url, resp.StatusCode, string(body))
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func compileProtos(localDir string) error {
	// Compila tutti i .proto in localDir (e sottocartelle)
	files := []string{}
	filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".proto") {
			files = append(files, path)
		}
		return nil
	})
	if len(files) == 0 {
		return fmt.Errorf("nessun .proto trovato in %s", localDir)
	}
	args := []string{
		"--go_out=paths=source_relative:.",
	}
	args = append(args, files...)
	cmd := exec.Command("protoc", args...)
	cmd.Dir = localDir
	var out bytes.Buffer
	cmd.Stderr = &out
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("protoc error: %v\n%s", err, out.String())
	}
	return nil
}
