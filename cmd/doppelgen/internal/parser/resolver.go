// Package parser. resolver.go - provides TypeResolver, which maps Go import paths to
// local filesystem directories, determines whether a type is project-internal or
// third-party, and caches ParseResult objects to avoid redundant AST parsing
// during recursive generation.
//
// A type is considered "project-internal" if its full import path begins with the
// current Go module path (read from go.mod) and its directory exists locally under
// the module root. All other types (from the module cache / vendor / GOPATH) are
// treated as third-party and receive convention-function stubs instead of generated
// Clone() methods.
package parser

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// -------------------------------------------- Types --------------------------------------------

// TypeResolver resolves Go import paths to local directories, decides whether a
// type is project-internal or third-party, and maintains a parse cache so each
// package is parsed at most once per generation run.
//
// TypeResolver is safe for concurrent use.
type TypeResolver struct {
	ModulePath string // ModulePath is the module declaration from go.mod (e.g. "github.com/seyallius/doppel").
	ModuleRoot string // ModuleRoot is the absolute filesystem path of the directory containing go.mod.

	mu    sync.RWMutex
	cache map[string]*ParseResult // keyed by import path
}

// -------------------------------------------- Constructor(s) --------------------------------------------

// NewTypeResolver creates a TypeResolver for the given module. Both modulePath and
// moduleRoot must be non-empty. Typically obtained from FindModuleRoot + ParseModulePath.
func NewTypeResolver(modulePath, moduleRoot string) *TypeResolver {
	return &TypeResolver{
		ModulePath: modulePath,
		ModuleRoot: moduleRoot,
		cache:      make(map[string]*ParseResult),
	}
}

// -------------------------------------------- Public Functions --------------------------------------------

// IsProjectInternal reports whether the given Go import path belongs to the current
// module. It checks two conditions:
//  1. The import path starts with the module path (e.g. "github.com/seyallius/doppel/manual").
//  2. The corresponding local directory actually exists on disk.
//
// The second check guards against accidental matches when module paths are prefixes
// of each other (e.g. "github.com/foo/bar" vs "github.com/foo/bar-extended").
func (r *TypeResolver) IsProjectInternal(importPath string) bool {
	if !strings.HasPrefix(importPath, r.ModulePath) {
		return false
	}

	// Require that the next character after the module prefix is "/" or end-of-string,
	// to avoid "github.com/foo/bar" matching "github.com/foo/bar-other".
	rest := strings.TrimPrefix(importPath, r.ModulePath)
	if rest != "" && !strings.HasPrefix(rest, "/") {
		return false
	}

	localDir, err := importPathToLocalDir(importPath, r.ModulePath, r.ModuleRoot)
	if err != nil {
		return false
	}

	info, err := os.Stat(localDir)
	return err == nil && info.IsDir()
}

// ImportPathToDir maps a project-internal import path to its absolute local
// directory. Returns an error if the import path is not under the module root.
func (r *TypeResolver) ImportPathToDir(importPath string) (string, error) {
	return importPathToLocalDir(importPath, r.ModulePath, r.ModuleRoot)
}

// GetOrParse returns a cached ParseResult for the given import path, or parses the
// package on the first call. Subsequent calls for the same import path return the
// cached result without re-parsing. Thread-safe via RWMutex.
//
// Returns an error if the import path cannot be resolved to a local directory, the
// directory does not exist, or parsing fails.
func (r *TypeResolver) GetOrParse(importPath, tagKey string) (*ParseResult, error) {
	// Fast path: read lock for cache hit.
	r.mu.RLock()
	if cached, ok := r.cache[importPath]; ok {
		r.mu.RUnlock()
		return cached, nil
	}
	r.mu.RUnlock()

	// Resolve the local directory.
	localDir, err := r.ImportPathToDir(importPath)
	if err != nil {
		return nil, fmt.Errorf("resolve import path %q to local dir: %w", importPath, err)
	}

	if _, statErr := os.Stat(localDir); statErr != nil {
		return nil, fmt.Errorf("local dir for %q does not exist (%s): %w", importPath, localDir, statErr)
	}

	// Parse the package. Use ParsePackageWithResolver so it inherits module awareness.
	result, err := ParsePackageWithResolver(localDir, tagKey, r)
	if err != nil {
		return nil, fmt.Errorf("parse package at %q (%s): %w", importPath, localDir, err)
	}

	// Store in cache (write lock).
	r.mu.Lock()
	r.cache[importPath] = result
	r.mu.Unlock()

	return result, nil
}

// WarnIfMissing emits a stderr warning when a project-internal import path cannot be
// resolved — e.g. when a type has been renamed or a go.mod dependency is missing.
// This is a best-effort diagnostic; it does not halt generation.
func (r *TypeResolver) WarnIfMissing(importPath string) {
	if r.IsProjectInternal(importPath) {
		return
	}

	// Only warn for paths that look like they should be internal (same org prefix etc.)
	if strings.HasPrefix(importPath, strings.Split(r.ModulePath, "/")[0]) {
		_, _ = fmt.Fprintf(os.Stderr,
			"doppelgen: warning: import path %q looks internal but is not under module %q — treating as third-party\n",
			importPath, r.ModulePath)
	}
}
