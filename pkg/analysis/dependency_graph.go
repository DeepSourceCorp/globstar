package analysis

import (
	"fmt"
	"path/filepath"
)

type DependencyGraph struct {
    dependencies map[string][]string // file -> dependencies
    dependents   map[string][]string // file -> files that depend on it
}

func NewDependencyGraph() *DependencyGraph {
    dg := &DependencyGraph{
        dependencies: map[string][]string{},
        dependents:   map[string][]string{},
    }

    return dg
}

func (dg *DependencyGraph) AddDependency(file, dependency string) {
    file = filepath.Clean(file)
    dependency = filepath.Clean(dependency)

    if file == dependency {
        return
    }

    for _, dep := range dg.dependencies[file] {
        if dep == dependency {
            return
        }
    }

    dg.dependencies[file] = append(dg.dependencies[file], dependency)
    dg.dependents[dependency] = append(dg.dependents[dependency], file)
}

// GetAffectedFiles returns all files affected by changes in the given files
func (dg *DependencyGraph) GetAffectedFiles(changedFiles []string) []string {
    seen := make(map[string]bool)
    result := make([]string, 0)

    var visit func(string)
    visit = func(file string) {
        if seen[file] {
            return
        }

        seen[file] = true
        result = append(result, file)

        // Find all files that depend on this file
        for _, dependent := range dg.dependents[file] {
            visit(dependent)
        }
    }

    // Start with the changed files
    for _, file := range changedFiles {
        visit(file)
    }

    return result
}

func (dg *DependencyGraph) GetFileDependencies(path string) error {
    res, ok := dg.dependencies[path]
    if !ok {
        return fmt.Errorf("No dependencies for file: %v", path)
    }
    fmt.Println(res)
    return nil
}

// PrintGraph prints a text representation of the dependency graph
func (dg *DependencyGraph) PrintGraph() {
    fmt.Println("Dependency Graph:")
    fmt.Println("=================")
    
    // Print files and their dependencies
    fmt.Println("\nFiles and their dependencies:")
    for file, deps := range dg.dependencies {
        if len(deps) > 0 {
            fmt.Printf("%s depends on:\n", file)
            for _, dep := range deps {
                fmt.Printf("  - %s\n", dep)
            }
            fmt.Println()
        }
    }
    
    // Print dependencies and their dependents
    fmt.Println("\nDependencies and their dependents:")
    for dep, files := range dg.dependents {
        if len(files) > 0 {
            fmt.Printf("%s is depended on by:\n", dep)
            for _, file := range files {
                fmt.Printf("  - %s\n", file)
            }
            fmt.Println()
        }
    }
}
