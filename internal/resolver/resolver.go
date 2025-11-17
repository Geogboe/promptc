package resolver

import (
	"fmt"
	"strings"

	"github.com/Geogboe/promptc/internal/library"
)

// Resolver resolves imports recursively with cycle detection
type Resolver struct {
	libraryManager *library.Manager
	visited        map[string]bool
	exclusions     map[string]bool
	resolved       map[string]string
}

// NewResolver creates a new import resolver
func NewResolver(libraryManager *library.Manager) *Resolver {
	return &Resolver{
		libraryManager: libraryManager,
		visited:        make(map[string]bool),
		exclusions:     make(map[string]bool),
		resolved:       make(map[string]string),
	}
}

// Resolve resolves all imports recursively
func (r *Resolver) Resolve(imports []string) (string, error) {
	// Reset state
	r.visited = make(map[string]bool)
	r.exclusions = make(map[string]bool)
	r.resolved = make(map[string]string)

	// First pass: collect exclusions
	for _, imp := range imports {
		if strings.HasPrefix(imp, "!") {
			r.exclusions[imp[1:]] = true
		}
	}

	// Second pass: resolve imports
	var contentParts []string
	for _, imp := range imports {
		if !strings.HasPrefix(imp, "!") {
			content, err := r.resolveRecursive(imp)
			if err != nil {
				return "", err
			}
			if content != "" {
				contentParts = append(contentParts, content)
			}
		}
	}

	return strings.Join(contentParts, "\n\n"), nil
}

func (r *Resolver) resolveRecursive(importName string) (string, error) {
	// Check if excluded
	if r.exclusions[importName] {
		return "", nil
	}

	// Check if already visited (cycle detection)
	if r.visited[importName] {
		return "", nil
	}

	// Check if already resolved
	if content, ok := r.resolved[importName]; ok {
		return content, nil
	}

	r.visited[importName] = true

	// Load the content
	content, err := r.libraryManager.Resolve(importName)
	if err != nil {
		return "", fmt.Errorf("failed to resolve import '%s': %w", importName, err)
	}

	// Store and return
	r.resolved[importName] = content
	return content, nil
}

// GetResolutionOrder returns the order in which imports were resolved
func (r *Resolver) GetResolutionOrder() []string {
	order := make([]string, 0, len(r.visited))
	for imp := range r.visited {
		order = append(order, imp)
	}
	return order
}
