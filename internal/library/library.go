package library

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed defaults/**/*.prompt
var defaultLibraries embed.FS

var (
	// ErrPathTraversal is returned when a path traversal attempt is detected
	ErrPathTraversal = errors.New("path traversal detected")
	// ErrInvalidImportName is returned when an import name is invalid
	ErrInvalidImportName = errors.New("invalid import name")
)

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

	// Clean the project directory path
	projectDir = filepath.Clean(projectDir)

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
		paths = append(paths, filepath.Clean(projectPrompts))
	}

	// 2. Global user prompts
	homeDir, err := os.UserHomeDir()
	if err == nil {
		homePrompts := filepath.Join(homeDir, ".prompts")
		if _, err := os.Stat(homePrompts); err == nil {
			paths = append(paths, filepath.Clean(homePrompts))
		}
	}

	// 3. Built-in defaults (embedded)
	paths = append(paths, "defaults")

	return paths
}

// Resolve resolves an import name to file content
func (m *Manager) Resolve(importName string) (string, error) {
	// Validate import name to prevent path traversal
	if err := validateImportName(importName); err != nil {
		return "", fmt.Errorf("invalid import name '%s': %w", importName, err)
	}

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
			// Try filesystem with security checks
			content, err := m.loadFromFilesystem(searchPath, relativePath)
			if err == nil {
				return content, nil
			}
		}
	}

	return "", fmt.Errorf("cannot resolve import '%s'. Searched in: %s",
		importName, strings.Join(m.SearchPaths, ", "))
}

// validateImportName validates that an import name doesn't contain path traversal sequences
func validateImportName(importName string) error {
	// Check for empty
	if importName == "" {
		return ErrInvalidImportName
	}

	// Check for path traversal attempts
	if strings.Contains(importName, "..") {
		return ErrPathTraversal
	}

	// Check for absolute paths
	if strings.HasPrefix(importName, "/") || strings.HasPrefix(importName, "\\") {
		return ErrPathTraversal
	}

	// Check for drive letters (Windows)
	if len(importName) >= 2 && importName[1] == ':' {
		return ErrPathTraversal
	}

	// Only allow alphanumeric, dots, and underscores
	for _, c := range importName {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '.' || c == '_' || c == '-') {
			return ErrInvalidImportName
		}
	}

	return nil
}

// loadFromFilesystem loads a prompt file from the filesystem with security checks
func (m *Manager) loadFromFilesystem(searchPath, relativePath string) (string, error) {
	// Build the full path
	fullPath := filepath.Join(searchPath, relativePath)

	// Clean the path to resolve any .. or . sequences
	fullPath = filepath.Clean(fullPath)
	cleanSearchPath := filepath.Clean(searchPath)

	// Verify the resolved path is still within the search path (prevent path traversal)
	if !strings.HasPrefix(fullPath, cleanSearchPath) {
		return "", fmt.Errorf("%w: attempted to access path outside allowed directory", ErrPathTraversal)
	}

	// Check if it's a symlink
	fileInfo, err := os.Lstat(fullPath)
	if err != nil {
		return "", err
	}

	// If it's a symlink, resolve and verify it's within allowed directory
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		resolvedPath, err := filepath.EvalSymlinks(fullPath)
		if err != nil {
			return "", fmt.Errorf("failed to resolve symlink: %w", err)
		}

		resolvedPath = filepath.Clean(resolvedPath)

		// Verify symlink target is within allowed directory
		if !strings.HasPrefix(resolvedPath, cleanSearchPath) {
			return "", fmt.Errorf("%w: symlink points outside allowed directory", ErrPathTraversal)
		}

		fullPath = resolvedPath
	}

	// Read file with size limit to prevent memory exhaustion
	const maxFileSize = 10 * 1024 * 1024 // 10MB limit
	fileInfo, err = os.Stat(fullPath)
	if err != nil {
		return "", err
	}

	if fileInfo.Size() > maxFileSize {
		return "", fmt.Errorf("file too large: %d bytes (max %d bytes)", fileInfo.Size(), maxFileSize)
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func (m *Manager) loadEmbedded(relativePath string) (string, error) {
	path := filepath.Join("defaults", relativePath)
	// Embedded FS is already safe from path traversal
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
