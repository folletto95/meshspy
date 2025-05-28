// ghdownloader/plugin.go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
)

// Content mappa la risposta GitHub API
type Content struct {
    Name        string `json:"name"`
    Path        string `json:"path"`
    Type        string `json:"type"`         // "file" o "dir"
    DownloadURL string `json:"download_url"` // solo per file
}

// DownloadProtos è il simbolo esportato dal plugin.
// owner, repo, path, ref(tag), out(dest), token (opzionale)
func DownloadProtos(owner, repo, path, ref, out, token string) error {
    client := &http.Client{}
    auth := ""
    if token != "" {
        auth = "token " + token
    }
    return downloadDir(client, auth, owner, repo, path, ref, out)
}

func downloadDir(client *http.Client, auth, owner, repo, dirPath, ref, localDir string) error {
    apiURL := fmt.Sprintf(
        "https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
        owner, repo, dirPath, ref,
    )
    req, _ := http.NewRequest("GET", apiURL, nil)
    if auth != "" {
        req.Header.Set("Authorization", auth)
    }
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, body)
    }

    var items []Content
    if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
        return err
    }
    if err := os.MkdirAll(localDir, 0755); err != nil {
        return err
    }

    for _, item := range items {
        dst := filepath.Join(localDir, item.Name)
        if item.Type == "file" {
            if err := downloadFile(client, auth, item.DownloadURL, dst); err != nil {
                return err
            }
        } else if item.Type == "dir" {
            if err := downloadDir(client, auth, owner, repo, item.Path, ref, dst); err != nil {
                return err
            }
        }
    }
    return nil
}

func downloadFile(client *http.Client, auth, url, dest string) error {
    req, _ := http.NewRequest("GET", url, nil)
    if auth != "" {
        req.Header.Set("Authorization", auth)
    }
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("download %s errore %d: %s", url, resp.StatusCode, body)
    }
    f, err := os.Create(dest)
    if err != nil {
        return err
    }
    defer f.Close()
    _, err = io.Copy(f, resp.Body)
    return err
}
