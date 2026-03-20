package fetcher

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// CacheDir returns the default cache directory: ~/.cache/promptc/resources/
func CacheDir() string {
	if runtime.GOOS == "windows" {
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			return filepath.Join(local, "promptc", "resources")
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".cache", "promptc", "resources")
	}
	return filepath.Join(os.TempDir(), "promptc", "resources")
}

// CacheKey returns the SHA-256 cache key for a resource.
// For URL resources: hash of the URL string.
// For git resources: hash of "<git-url>@<ref>" (ref defaults to "HEAD").
func CacheKey(r Resource) string {
	var input string
	if r.URL != "" {
		input = r.URL
	} else {
		ref := r.Ref
		if ref == "" {
			ref = "HEAD"
		}
		input = r.Git + "@" + ref
	}
	if r.Path != "" {
		input += "#" + r.Path
	}
	sum := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", sum)
}

// CachePath returns the directory where a fetched resource would be cached.
func CachePath(r Resource) string {
	return filepath.Join(CacheDir(), CacheKey(r))
}

// IsCached reports whether a resource is already in the cache.
func IsCached(r Resource) bool {
	_, err := os.Stat(CachePath(r))
	return err == nil
}

// CopyFromCache copies the cached resource into destDir.
// Returns an error if the cache entry does not exist.
func CopyFromCache(r Resource, destDir string) error {
	src := CachePath(r)
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("cache miss for %q", r.Name)
	}

	dest := filepath.Join(destDir, r.Name)
	if info.IsDir() {
		return copyDir(src, dest)
	}
	return copyFile(src, dest+".md")
}

// StoreInCache copies a fetched resource from destDir into the cache.
func StoreInCache(r Resource, destDir string) error {
	cachePath := CachePath(r)
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return err
	}

	src := filepath.Join(destDir, r.Name)
	if r.URL != "" {
		src += ".md"
	}

	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDir(src, cachePath)
	}
	return copyFile(src, cachePath)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(path, target)
	})
}
