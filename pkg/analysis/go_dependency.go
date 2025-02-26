package analysis

import (
	"os"
	"path/filepath"
	"strings"
)


func FindGoFilesInPackage(packageDir string) ([]string, error) {
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