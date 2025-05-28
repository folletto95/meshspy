// ghdownloader/plugin.go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "sort"
    "strconv"
    "strings"
)

// Content mappa la risposta GitHub API
type Content struct {
    Name        string `json:"name"`
    Path        string `json:"path"`
    Type        string `json:"type"`         // "file" o "dir"
    DownloadURL string `json:"download_url"` // solo per file
}

type Tag struct {
    Name string `json:"name"`
}

// ===== ESPORTATA =====
// Chiama questa dal main:
//    DownloadAllProtos(token string)
// Scarica TUTTE le versioni >= v2.0.14 e genera i binding Go
func DownloadAllProtos(token string) error {
    owner := "meshtastic"
    repo := "protobufs"
    repoPath := "meshtastic"
    minTag := "v2.0.14"
    protoBase := "./meshtastic-proto"
    pbBase := "./pb"

    tags, err := getTags(owner, repo, minTag, token)
    if err != nil {
        return err
    }
    log.Printf("➡️  Trovati tag: %v", tags)
    for _, tag := range tags {
        protoOut := filepath.Join(protoBase, tag, repoPath)
        pbOut := filepath.Join(pbBase, tag, repoPath)
        log.Printf("⬇️  [%s] Scarico proto...", tag)
        if err := downloadDir(token, owner, repo, repoPath, tag, protoOut); err != nil {
            log.Printf("❌ [%s] Errore download: %v", tag, err)
            continue
        }
        log.Printf("✔️  [%s] Proto scaricati in %s", tag, protoOut)
        if err := os.MkdirAll(pbOut, 0755); err != nil {
            log.Printf("❌ [%s] Errore mkdir: %v", tag, err)
            continue
        }
        protoFiles, _ := filepath.Glob(filepath.Join(protoOut, "*.proto"))
        if len(protoFiles) == 0 {
            log.Printf("⚠️  [%s] Nessun .proto trovato in %s", tag, protoOut)
            continue
        }
        log.Printf("🛠️  [%s] Compilo .pb.go...", tag)
        args := []string{
            "--go_out=" + pbOut,
            "--go_opt=paths=source_relative",
            "--proto_path=" + protoOut,
        }
        args = append(args, protoFiles...)
        cmd := exec.Command("protoc", args...)
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
            log.Printf("❌ [%s] Errore protoc: %v", tag, err)
        } else {
            log.Printf("✅ [%s] Binding Go generati in %s", tag, pbOut)
        }
    }
    return nil
}

// ====== GESTIONE TAG GITHUB API ======

func getTags(owner, repo, minTag, token string) ([]string, error) {
    apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/tags", owner, repo)
    req, err := http.NewRequest("GET", apiURL, nil)
    if err != nil {
        return nil, err
    }
    if token != "" {
        req.Header.Set("Authorization", "token "+token)
    }
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("GitHub API tags error %d: %s", resp.StatusCode, body)
    }
    var tags []Tag
    if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
        return nil, err
    }
    // Filtra >= minTag (es: v2.0.14)
    min := parseVer(minTag)
    var out []string
    for _, t := range tags {
        if strings.HasPrefix(t.Name, "v") && compareVer(parseVer(t.Name), min) >= 0 {
            out = append(out, t.Name)
        }
    }
    // Ordina (crescente)
    sort.Slice(out, func(i, j int) bool { return compareVer(parseVer(out[i]), parseVer(out[j])) < 0 })
    return out, nil
}

// ======= DOWNLOAD .proto ========

func downloadDir(token, owner, repo, dirPath, ref, localDir string) error {
    apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", owner, repo, dirPath, ref)
    req, _ := http.NewRequest("GET", apiURL, nil)
    if token != "" {
        req.Header.Set("Authorization", "token "+token)
    }
    resp, err := http.DefaultClient.Do(req)
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
            if err := downloadFile(token, item.DownloadURL, dst); err != nil {
                return err
            }
        } else if item.Type == "dir" {
            if err := downloadDir(token, owner, repo, item.Path, ref, dst); err != nil {
                return err
            }
        }
    }
    return nil
}

func downloadFile(token, url, dest string) error {
    req, _ := http.NewRequest("GET", url, nil)
    if token != "" {
        req.Header.Set("Authorization", "token "+token)
    }
    resp, err := http.DefaultClient.Do(req)
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

// ====== VERSIONI =======

var versionRegexp = regexp.MustCompile(`v(\d+)\.(\d+)\.(\d+)`)
func parseVer(s string) [3]int {
    m := versionRegexp.FindStringSubmatch(s)
    if len(m) != 4 {
        return [3]int{0, 0, 0}
    }
    return [3]int{atoi(m[1]), atoi(m[2]), atoi(m[3])}
}
func compareVer(a, b [3]int) int {
    for i := 0; i < 3; i++ {
        if a[i] < b[i] {
            return -1
        }
        if a[i] > b[i] {
            return 1
        }
    }
    return 0
}
func atoi(s string) int { i, _ := strconv.Atoi(s); return i }
