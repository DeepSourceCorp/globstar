package checkers

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/DeepSourceCorp/globstar/pkg/analysis"
)

//go:embed *.yaml *.yml
var builtinCheckers embed.FS

func LoadBuiltinCheckers() (map[analysis.Language][]analysis.YmlRule, error) {
	rulesMap := make(map[analysis.Language][]analysis.YmlRule)
	d, err := builtinCheckers.ReadDir(".")
	fmt.Printf("d: %v, err: %s\n", d, err)
	err = fs.WalkDir(builtinCheckers, ".", func(path string, d fs.DirEntry, err error) error {
		fmt.Println(path)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		if d.IsDir() {
			fmt.Println("is dir", path)
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
