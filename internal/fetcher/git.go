package fetcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	getter "github.com/hashicorp/go-getter"
)

// FetchGit clones a git repository into <destDir>/<resource.Name>/ using go-getter.
// go-getter handles shallow clones, branch/tag/commit refs, and subdirectory extraction.
func FetchGit(ctx context.Context, resource Resource, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", destDir, err)
	}

	src := buildGetterURL(resource)
	dst := filepath.Join(destDir, resource.Name)

	client := &getter.Client{
		Ctx:  ctx,
		Src:  src,
		Dst:  dst,
		Mode: getter.ClientModeDir,
	}
	return client.Get()
}

// buildGetterURL constructs a go-getter URL for a git resource.
// Format: git::<repo-url>//[subdir][?ref=<ref>]
func buildGetterURL(r Resource) string {
	url := "git::" + r.Git
	if r.Path != "" {
		url += "//" + r.Path
	}

	var params []string
	if r.Ref != "" {
		params = append(params, "ref="+r.Ref)
	}
	// Request a depth-1 shallow clone for speed
	params = append(params, "depth=1")

	if len(params) > 0 {
		url += "?"
		for i, p := range params {
			if i > 0 {
				url += "&"
			}
			url += p
		}
	}
	return url
}
