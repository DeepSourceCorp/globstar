package checkers

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"

	goAnalysis "globstar.dev/analysis"
	"globstar.dev/checkers/golang"
	"globstar.dev/checkers/javascript"
	"globstar.dev/pkg/analysis"
)

//go:embed **/*.y*ml
var builtinCheckers embed.FS

func LoadYamlRules() (map[analysis.Language][]analysis.YmlRule, error) {
	rulesMap := make(map[analysis.Language][]analysis.YmlRule)
	err := fs.WalkDir(builtinCheckers, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		fileExt := filepath.Ext(path)
		isYamlFile := fileExt == ".yaml" || fileExt == ".yml"
		if !isYamlFile {
			return nil
		}

		fileContent, err := builtinCheckers.ReadFile(path)
		if err != nil {
			return nil
		}

		patternRule, err := analysis.ReadFromBytes(fileContent)
		if err != nil {
			return fmt.Errorf("invalid rule '%s': %s", d.Name(), err.Error())
		}

		lang := patternRule.Language()
		rulesMap[lang] = append(rulesMap[lang], patternRule)
		return nil
	})
	return rulesMap, err
}

func LoadGoRules() []*goAnalysis.Analyzer {
	return []*goAnalysis.Analyzer{
		&javascript.NoDoubleEq,
		&golang.WeakScryptCost,
	}
}
