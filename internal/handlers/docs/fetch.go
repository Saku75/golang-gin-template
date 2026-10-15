//go:build ignore

// Fetches the Stoplight Elements files that the docs page serves into assets/.
// They are downloaded as the npm package from NPM_REGISTRY, which defaults to
// the public registry, and verified against the registry's integrity hash.
// Run with go generate; it does nothing when the files are already present.
package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	pkg     = "@stoplight/elements"
	version = "9.0.15"
	dir     = "assets"
)

var files = []string{"web-components.min.js", "styles.min.css"}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fetch docs assets:", err)
		os.Exit(1)
	}
}

func run() error {
	if present() {
		return nil
	}

	registry := strings.TrimSuffix(os.Getenv("NPM_REGISTRY"), "/")
	if registry == "" {
		registry = "https://registry.npmjs.org"
	}
	client := &http.Client{Timeout: 2 * time.Minute}
	fmt.Printf("fetching %s@%s from %s\n", pkg, version, registry)

	var meta struct {
		Dist struct {
			Tarball   string `json:"tarball"`
			Integrity string `json:"integrity"`
		} `json:"dist"`
	}
	body, err := get(client, registry+"/"+url.PathEscape(pkg)+"/"+version)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, &meta); err != nil {
		return fmt.Errorf("decode package metadata: %w", err)
	}
	if meta.Dist.Tarball == "" {
		return errors.New("package metadata has no tarball URL")
	}

	tgz, err := get(client, meta.Dist.Tarball)
	if err != nil {
		return err
	}
	if err := verify(tgz, meta.Dist.Integrity); err != nil {
		return err
	}
	return extract(tgz)
}

func present() bool {
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			return false
		}
	}
	return true
}

func get(client *http.Client, u string) ([]byte, error) {
	resp, err := client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", u, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func verify(data []byte, integrity string) error {
	want, ok := strings.CutPrefix(integrity, "sha512-")
	if !ok {
		fmt.Println("warning: registry gave no sha512 integrity hash, skipping verification")
		return nil
	}
	sum := sha512.Sum512(data)
	if got := base64.StdEncoding.EncodeToString(sum[:]); got != want {
		return fmt.Errorf("tarball sha512 %s does not match registry integrity %s", got, want)
	}
	return nil
}

func extract(tgz []byte) error {
	gz, err := gzip.NewReader(bytes.NewReader(tgz))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	found := map[string]bool{}
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		name := strings.TrimPrefix(hdr.Name, "package/")
		for _, f := range files {
			if name != f {
				continue
			}
			data, err := io.ReadAll(tr)
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(dir, f), data, 0o644); err != nil {
				return err
			}
			found[f] = true
		}
	}
	for _, f := range files {
		if !found[f] {
			return fmt.Errorf("%s not found in %s@%s", f, pkg, version)
		}
	}
	return nil
}
