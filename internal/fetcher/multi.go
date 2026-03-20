package fetcher

import (
	"context"
	"fmt"
	"sync"

	"github.com/Geogboe/promptc/internal/progress"
)

// getterMu serializes all go-getter calls.
// go-getter's Configure() mutates globally registered getter singletons
// (FileGetter, GitGetter, etc.) via SetClient(), making concurrent Get()
// calls unsafe. Fetches are I/O-bound so this has minimal practical impact.
var getterMu sync.Mutex

// FetchAll fetches all resources concurrently into destDir.
// Collects all errors so partial failures are visible.
func FetchAll(ctx context.Context, resources []Resource, destDir string) []error {
	return fetchAll(ctx, resources, destDir, false)
}

// FetchAllCached fetches all resources, using the cache when available.
// Set noCache=true to force re-fetch even when cached.
func FetchAllCached(ctx context.Context, resources []Resource, destDir string, noCache bool) []error {
	return fetchAll(ctx, resources, destDir, noCache)
}

func fetchAll(ctx context.Context, resources []Resource, destDir string, noCache bool) []error {
	if len(resources) == 0 {
		return nil
	}

	var (
		mu     sync.Mutex
		errors []error
		wg     sync.WaitGroup
	)

	for _, r := range resources {
		wg.Add(1)
		go func(res Resource) {
			defer wg.Done()

			if !noCache && IsCached(res) {
				progress.Step("using cached %s", res.Name)
				if err := CopyFromCache(res, destDir); err != nil {
					progress.Fail("cache read %s: %v", res.Name, err)
					mu.Lock()
					errors = append(errors, fmt.Errorf("resource %q cache: %w", res.Name, err))
					mu.Unlock()
				} else {
					progress.Done("loaded %s from cache", res.Name)
				}
				return
			}

			progress.Step("fetching %s (%s)", res.Name, sourceLabel(res))

			var err error
			getterMu.Lock()
			if res.URL != "" {
				err = FetchHTTP(ctx, res, destDir)
			} else {
				err = FetchGit(ctx, res, destDir)
			}
			getterMu.Unlock()

			if err != nil {
				progress.Fail("fetch %s: %v", res.Name, err)
				mu.Lock()
				errors = append(errors, fmt.Errorf("resource %q: %w", res.Name, err))
				mu.Unlock()
				return
			}

			progress.Done("fetched %s", res.Name)

			if cacheErr := StoreInCache(res, destDir); cacheErr != nil {
				progress.Fail("cache write %s: %v (non-fatal)", res.Name, cacheErr)
			}
		}(r)
	}

	wg.Wait()
	return errors
}

func sourceLabel(r Resource) string {
	if r.URL != "" {
		return "URL"
	}
	return "git"
}
