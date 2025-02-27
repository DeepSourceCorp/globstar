package analysis

import (
	"os"
	"path/filepath"
	"strings"
)


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