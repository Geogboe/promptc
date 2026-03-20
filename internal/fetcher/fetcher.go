// Package fetcher fetches external resources (HTTP URLs and git repos)
// and stores them in a local directory for use by the build pipeline.
package fetcher

// Resource describes a single external resource to fetch.
type Resource struct {
	Name string // destination name (used as filename or directory)
	URL  string // HTTP/HTTPS URL (mutually exclusive with Git)
	Git  string // git repository URL (mutually exclusive with URL)
	Ref  string // branch/tag/commit (git only; default: default branch)
	Path string // subdirectory within the repo to copy (git only; optional)
}
