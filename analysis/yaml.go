package analysis

import (
	"fmt"
	"os"
	"strings"

	"github.com/gobwas/glob"
	sitter "github.com/smacker/go-tree-sitter"
	"gopkg.in/yaml.v3"
)

// To get a node back from a tree-sitter query, it *must* have a capture name.
// So: (call_expression) will match nothing, but (call_expression) @some_key
// will match all call expressions.
// For filtering patterns with clauses in the yaml file, like:
// filters:
//   - pattern-inside: (call_expression)
//   - pattern-not-inside: (catch_block)
//
// We need to append a key name at the end of the pattern written by the user.
// This is the key that we will use.
const filterPatternKey = "__filter__key__"

type filterYaml struct {
	PatternInside    string `yaml:"pattern-inside,omitempty"`
	PatternNotInside string `yaml:"pattern-not-inside,omitempty"`
}

type pathFilterYaml struct {
	Exclude []string `yaml:"exclude,omitempty"`
	Include []string `yaml:"include,omitempty"`
}

// NodeFilter is a filter that can be applied to a PatternChecker to restrict
// the nodes that the checker is applied to.
// The checker is only applied to nodes that have a parent matching (or not matching) the query.
type NodeFilter struct {
	query       *sitter.Query
	shouldMatch bool
}

// PathFilter is a glob that can be applied to a PatternChecker to restrict
// the files that the checker is applied to.
type PathFilter struct {
	ExcludeGlobs []glob.Glob
	IncludeGlobs []glob.Glob
}

type Yaml struct {
	Language    string          `yaml:"language"`
	Code        string          `yaml:"name"`
	Message     string          `yaml:"message"`
	Category    Category        `yaml:"category"`
	Severity    Severity        `yaml:"severity"`
	Pattern     string          `yaml:"pattern"`
	Patterns    []string        `yaml:"patterns"`
	Description string          `yaml:"description"`
	Exclude     []string        `yaml:"exclude,omitempty"`
	Include     []string        `yaml:"include,omitempty"`
	Filters     []filterYaml    `yaml:"filters,omitempty"`
	PathFilter  *pathFilterYaml `yaml:"path_filter,omitempty"`
}

type YamlAnalyzer struct {
	Analyzer   *Analyzer
	Patterns   []*sitter.Query
	NodeFilter []NodeFilter
	PathFilter *PathFilter
	Message    string
}

// ReadFromFile reads a pattern checker definition from a YAML config file.
func ReadFromFile(filePath string) (Analyzer, YamlAnalyzer, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return Analyzer{}, YamlAnalyzer{}, err
	}

	return ReadFromBytes(fileContent)
}

// ReadFromBytes reads a pattern checker definition from bytes array
func ReadFromBytes(fileContent []byte) (Analyzer, YamlAnalyzer, error) {
	var checker Yaml
	if err := yaml.Unmarshal(fileContent, &checker); err != nil {
		return Analyzer{}, YamlAnalyzer{}, err
	}

	lang, code, message, err := verifyChecker(checker)
	if err != nil {
		return Analyzer{}, YamlAnalyzer{}, err
	}

	var patterns []*sitter.Query
	if checker.Pattern != "" {
		pattern, err := sitter.NewQuery([]byte(checker.Pattern), lang.Grammar())
		if err != nil {
			return Analyzer{}, YamlAnalyzer{}, err
		}
		patterns = append(patterns, pattern)
	} else if len(checker.Patterns) > 0 {
		for _, patternStr := range checker.Patterns {
			pattern, err := sitter.NewQuery([]byte(patternStr), lang.Grammar())
			if err != nil {
				return Analyzer{}, YamlAnalyzer{}, fmt.Errorf("invalid tree-sitter query in one of the patterns")
			}
			patterns = append(patterns, pattern)
		}
	} else {
		return Analyzer{}, YamlAnalyzer{}, fmt.Errorf("no pattern provided in checker '%s'", code)
	}

	if checker.Pattern != "" && len(checker.Patterns) > 0 {
		return Analyzer{}, YamlAnalyzer{}, fmt.Errorf("only one of 'pattern' or 'patterns' can be provided in a checker definition")
	}

	// include and exclude patterns
	var pathFilter *PathFilter
	if checker.Exclude != nil || checker.Include != nil {
		pathFilter = &PathFilter{
			ExcludeGlobs: make([]glob.Glob, 0, len(checker.Exclude)),
			IncludeGlobs: make([]glob.Glob, 0, len(checker.Include)),
		}

		for _, exclude := range checker.Exclude {
			g, err := glob.Compile(exclude)
			if err != nil {
				return Analyzer{}, YamlAnalyzer{}, fmt.Errorf("invalid exclude pattern in yaml checker")
			}
			pathFilter.ExcludeGlobs = append(pathFilter.ExcludeGlobs, g)
		}

		for _, include := range checker.Include {
			g, err := glob.Compile(include)
			if err != nil {
				return Analyzer{}, YamlAnalyzer{}, fmt.Errorf("invalid include pattern in yaml checker")
			}
			pathFilter.IncludeGlobs = append(pathFilter.IncludeGlobs, g)
		}
	}

	// node filters
	var filters []NodeFilter
	if checker.Filters != nil {
		for _, filter := range checker.Filters {
			if filter.PatternInside != "" {
				queryStr := filter.PatternInside + " @" + filterPatternKey
				query, err := sitter.NewQuery([]byte(queryStr), lang.Grammar())
				if err != nil {
					return Analyzer{}, YamlAnalyzer{}, fmt.Errorf("invalid tree-sitter pattern inside 'pattern-inside' field")
				}

				filters = append(filters, NodeFilter{
					query:       query,
					shouldMatch: true,
				})
			}

			if filter.PatternNotInside != "" {
				queryStr := filter.PatternNotInside + " @" + filterPatternKey
				query, err := sitter.NewQuery([]byte(queryStr), lang.Grammar())
				if err != nil {
					return Analyzer{}, YamlAnalyzer{}, fmt.Errorf("invalid tree-sitter pattern inside 'pattern-not-inside' field")
				}

				filters = append(filters, NodeFilter{
					query:       query,
					shouldMatch: false,
				})
			}
		}
	}

	patternChecker := Analyzer{
		Name:        code,
		Language:    lang,
		Description: checker.Description,
		Category:    checker.Category,
		Severity:    checker.Severity,
	}

	yamlAnalyzer := &YamlAnalyzer{
		Analyzer:   &patternChecker,
		Patterns:   patterns,
		NodeFilter: filters,
		PathFilter: pathFilter,
		Message:    message,
	}

	patternChecker.Run = RunYamlAnalyzer(yamlAnalyzer)
	return patternChecker, *yamlAnalyzer, nil
}

func RunYamlAnalyzer(YamlAnalyzer *YamlAnalyzer) func(pass *Pass) (any, error) {
	return func(pass *Pass) (any, error) {
		queries := YamlAnalyzer.Patterns
		for _, query := range queries {
			qc := sitter.NewQueryCursor()
			defer qc.Close()
			qc.Exec(query, pass.FileContext.Ast)
			for {
				m, ok := qc.NextMatch()
				if !ok {
					break
				}
				m = qc.FilterPredicates(m, pass.FileContext.Source)
				for _, capture := range m.Captures {
					captureName := query.CaptureNameForId(capture.Index)
					if captureName == pass.Analyzer.Name && YamlAnalyzer.runParentFilters(pass.FileContext.Source, capture.Node) {
						message := YamlAnalyzer.Message
						for _, capture := range m.Captures {
							captureName := query.CaptureNameForId(capture.Index)
							message = strings.ReplaceAll(message, "@"+captureName, capture.Node.Content(pass.FileContext.Source))
						}

						pass.Report(pass, capture.Node, message)
					}
				}

			}
		}
		return nil, nil
	}

}

func (ana *YamlAnalyzer) runParentFilters(source []byte, capture *sitter.Node) bool {
	filters := ana.NodeFilter
	if len(filters) == 0 {
		return true
	}

	for _, filter := range filters {
		shouldMatch := filter.shouldMatch
		nodeMatched := false

		for parent := capture.Parent(); parent != nil; parent = parent.Parent() {
			if filterMatchesParent(&filter, parent, source) {
				nodeMatched = true
				if !shouldMatch {
					return false
				} else {
					break
				}
			}
		}

		if !nodeMatched && shouldMatch {
			return false
		}
	}

	return true
}

func filterMatchesParent(filter *NodeFilter, parent *sitter.Node, source []byte) bool {
	qc := sitter.NewQueryCursor()
	defer qc.Close()

	qc.Exec(filter.query, parent)

	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}

		m = qc.FilterPredicates(m, source)
		for _, capture := range m.Captures {
			captureName := filter.query.CaptureNameForId(capture.Index)
			if captureName == filterPatternKey && capture.Node == parent {
				return true
			}
		}
	}

	return false
}

func verifyChecker(checker Yaml) (Language, string, string, error) {
	lang := DecodeLanguage(checker.Language)
	code := checker.Code
	msg := checker.Message

	if lang == LangUnknown {
		return lang, code, msg, fmt.Errorf("unknown language code: %v", lang)
	}

	if (code == "") || (msg == "") {
		return lang, code, msg, fmt.Errorf("missing necessary field in checker definition")
	}

	return lang, code, msg, nil
}
