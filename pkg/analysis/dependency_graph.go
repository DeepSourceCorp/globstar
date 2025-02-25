package analysis

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DependencyGraph struct {
    dependencies map[string][]string // file -> dependencies
    dependents   map[string][]string // file -> files that depend on it
    moduleName string
}

func NewDependencyGraph() *DependencyGraph {
    dg := &DependencyGraph{
        dependencies: map[string][]string{},
        dependents:   map[string][]string{},
    }

    return dg
}

func (dg *DependencyGraph) IsInternalDependency(dependency string, rootDir string, language Language) bool {
    if filepath.IsAbs(dependency) {
        relPath, err := filepath.Rel(rootDir, dependency)
        if err != nil {
            return false
        }

        return !filepath.IsAbs(relPath) && !strings.HasPrefix(relPath, "..")
    }

    // Language-specific checks
    switch language {
    case LangGo:
        //extract the moduleName
        var modulePrefix string
        goMod := filepath.Join(rootDir, "go.mod")
        content, err := os.ReadFile(goMod)
        if err != nil {
            return false
        }
        lines := strings.Split(string(content), "\n")
        for _, line := range lines {
            line = strings.TrimSpace(line)
            if strings.HasPrefix(line, "module "){
                modulePrefix = strings.TrimSpace(strings.TrimPrefix(line, "module"))
                dg.moduleName = modulePrefix
            }else{
               continue 
            }
            
        }

        return strings.HasPrefix(dependency, modulePrefix)
    case LangJs, LangTs:
        // For JS/TS, relative imports are internal
        return strings.HasPrefix(dependency, "./") || strings.HasPrefix(dependency, "../")
    case LangPy:
        // For Python, check if it's a local module
        // This is a simplified check
        return !strings.Contains(dependency, ".")
    }
    
    return false
}

func (dg *DependencyGraph) AddDependency(file, dependency string, rootDir string) {
    file = filepath.Clean(file)
    dependency = filepath.Clean(dependency)

    if file == dependency {
        return
    }
    language := LanguageFromFilePath(file)
    if !dg.IsInternalDependency(dependency, rootDir, language) {
        return
    }

    if language == LangGo && !filepath.IsAbs(dependency) {
        // Convert import path to directory path
        relPath := strings.TrimPrefix(dependency, dg.moduleName)
        // Remove leading slash if present
        relPath = strings.TrimPrefix(relPath, "/")
        packageDir := filepath.Join(rootDir, relPath)
        
        // Find all Go files in the package
        goFiles, err := findGoFilesInPackage(packageDir)
        if err == nil && len(goFiles) > 0 {
            // Add each Go file as a dependency
            for _, goFile := range goFiles {
                if file == goFile {
                    continue // Skip self-dependencies
                }
                
                // Check if this dependency already exists
                alreadyExists := false
                for _, dep := range dg.dependencies[file] {
                    if dep == goFile {
                        alreadyExists = true
                        break
                    }
                }
                
                if !alreadyExists {
                    dg.dependencies[file] = append(dg.dependencies[file], goFile)
                    dg.dependents[goFile] = append(dg.dependents[goFile], file)
                }
            }
            return
        }
        // If we couldn't find Go files, fall back to using the package directory
        dependency = packageDir
    }

    for _, dep := range dg.dependencies[file] {
        if dep == dependency {
            return
        }
    }

    // if LanguageFromFilePath(file) == LangGo{
    //     relPath := strings.TrimPrefix(dependency, dg.moduleName)
    //     fmt.Println(relPath)
    //     dependency = filepath.Join(rootDir,relPath)
    // }

    dg.dependencies[file] = append(dg.dependencies[file], dependency)
    dg.dependents[dependency] = append(dg.dependents[dependency], file)
}

func findGoFilesInPackage(packageDir string) ([]string, error) {
    var goFiles []string
    
    entries, err := os.ReadDir(packageDir)
    if err != nil {
        return nil, err
    }
    
    for _, entry := range entries {
        // Skip directories and non-Go files
        if entry.IsDir() {
            continue
        }
        
        name := entry.Name()
        if !strings.HasSuffix(name, ".go") {
            continue
        }
        
        // Skip test files
        if strings.HasSuffix(name, "_test.go") {
            continue
        }
        
        goFiles = append(goFiles, filepath.Join(packageDir, name))
    }
    
    return goFiles, nil
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

func (dg *DependencyGraph) GetFileDependencies(path string) ([]string,error) {
    res, ok := dg.dependencies[path]
    if !ok {
        return []string{}, fmt.Errorf("No dependencies for file: %v", path)
    }
    return res, nil
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
