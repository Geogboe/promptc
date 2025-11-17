package library

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed defaults/**/*.prompt
var defaultLibraries embed.FS

// Manager handles prompt library resolution and loading
type Manager struct {
	ProjectDir  string
	SearchPaths []string
}

// NewManager creates a new library manager
func NewManager(projectDir string) *Manager {
	if projectDir == "" {
		var err error
		projectDir, err = os.Getwd()
		if err != nil {
			projectDir = "."
		}
	}

	m := &Manager{
		ProjectDir: projectDir,
	}
	m.SearchPaths = m.getSearchPaths()
	return m
}

func (m *Manager) getSearchPaths() []string {
	var paths []string

	// 1. Project-local prompts
	projectPrompts := filepath.Join(m.ProjectDir, "prompts")
	if _, err := os.Stat(projectPrompts); err == nil {
		paths = append(paths, projectPrompts)
	}

	// 2. Global user prompts
	homeDir, err := os.UserHomeDir()
	if err == nil {
		homePrompts := filepath.Join(homeDir, ".prompts")
		if _, err := os.Stat(homePrompts); err == nil {
			paths = append(paths, homePrompts)
		}
	}

	// 3. Built-in defaults (embedded)
	paths = append(paths, "defaults")

	return paths
}

// Resolve resolves an import name to file content
func (m *Manager) Resolve(importName string) (string, error) {
	// Convert dot notation to path (patterns.rest_api -> patterns/rest_api.prompt)
	relativePath := strings.ReplaceAll(importName, ".", string(filepath.Separator)) + ".prompt"

	// Try each search path
	for _, searchPath := range m.SearchPaths {
		if searchPath == "defaults" {
			// Try embedded files
			content, err := m.loadEmbedded(relativePath)
			if err == nil {
				return content, nil
			}
		} else {
			// Try filesystem
			fullPath := filepath.Join(searchPath, relativePath)
			content, err := os.ReadFile(fullPath)
			if err == nil {
				return string(content), nil
			}
		}
	}

	return "", fmt.Errorf("cannot resolve import '%s'. Searched in: %s",
		importName, strings.Join(m.SearchPaths, ", "))
}

func (m *Manager) loadEmbedded(relativePath string) (string, error) {
	path := filepath.Join("defaults", relativePath)
	content, err := defaultLibraries.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// Library represents available libraries by source
type Library struct {
	Project  []string
	Global   []string
	BuiltIn  []string
}

// ListLibraries returns all available libraries organized by source
func (m *Manager) ListLibraries() *Library {
	lib := &Library{
		Project:  []string{},
		Global:   []string{},
		BuiltIn:  []string{},
	}

	for _, searchPath := range m.SearchPaths {
		if searchPath == "defaults" {
			lib.BuiltIn = m.scanEmbedded()
		} else {
			libs := m.scanDirectory(searchPath, "")
			if strings.Contains(searchPath, "prompts") && filepath.Dir(searchPath) == m.ProjectDir {
				lib.Project = libs
			} else {
				lib.Global = libs
			}
		}
	}

	return lib
}

func (m *Manager) scanDirectory(basePath, prefix string) []string {
	var libraries []string

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return libraries
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		if entry.IsDir() {
			subPrefix := name
			if prefix != "" {
				subPrefix = prefix + "." + name
			}
			libraries = append(libraries, m.scanDirectory(filepath.Join(basePath, name), subPrefix)...)
		} else if strings.HasSuffix(name, ".prompt") {
			libName := strings.TrimSuffix(name, ".prompt")
			if prefix != "" {
				libName = prefix + "." + libName
			}
			libraries = append(libraries, libName)
		}
	}

	sort.Strings(libraries)
	return libraries
}

func (m *Manager) scanEmbedded() []string {
	var libraries []string

	entries, err := defaultLibraries.ReadDir("defaults")
	if err != nil {
		return libraries
	}

	for _, entry := range entries {
		if entry.IsDir() {
			libraries = append(libraries, m.scanEmbeddedDir("defaults/"+entry.Name(), entry.Name())...)
		}
	}

	sort.Strings(libraries)
	return libraries
}

func (m *Manager) scanEmbeddedDir(path, prefix string) []string {
	var libraries []string

	entries, err := defaultLibraries.ReadDir(path)
	if err != nil {
		return libraries
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			subPrefix := prefix + "." + name
			libraries = append(libraries, m.scanEmbeddedDir(path+"/"+name, subPrefix)...)
		} else if strings.HasSuffix(name, ".prompt") {
			libName := prefix + "." + strings.TrimSuffix(name, ".prompt")
			libraries = append(libraries, libName)
		}
	}

	return libraries
}
