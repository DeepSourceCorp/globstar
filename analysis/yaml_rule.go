package analysis

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
	"github.com/gobwas/glob"
	sitter "github.com/smacker/go-tree-sitter"
)

const filterPatternKey = "__filter__key__"

type VisitFn func(checker Checker, analyzer *YamlAnalyzer, node *sitter.Node)

type Checker interface {
	NodeType() string
	GetLanguage() Language
	OnEnter() *VisitFn
	OnLeave() *VisitFn
}

type checkerImpl struct {
	nodeType string
	language Language
	onEnter  *VisitFn
	onLeave  *VisitFn
}

func (r *checkerImpl) NodeType() string      { return r.nodeType }
func (r *checkerImpl) GetLanguage() Language { return r.language }
func (r *checkerImpl) OnEnter() *VisitFn     { return r.onEnter }
func (r *checkerImpl) OnLeave() *VisitFn     { return r.onLeave }

func CreateChecker(nodeType string, language Language, onEnter, onLeave *VisitFn) Checker {
	return &checkerImpl{
		nodeType: nodeType,
		language: language,
		onEnter:  onEnter,
		onLeave:  onLeave,
	}
}

// ReadFromFile reads a pattern checker definition from a YAML config file.
func ReadFromFile(filePath string) (YamlChecker, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return ReadFromBytes(fileContent)
}

func DecodeLanguage(language string) Language {
	language = strings.ToLower(language)
	switch language {
	case "javascript", "js":
		return LangJs
	case "typescript", "ts":
		return LangTs
	case "jsx", "tsx":
		return LangTsx
	case "python", "py":
		return LangPy
	case "ocaml", "ml":
		return LangOCaml
	case "docker", "dockerfile":
		return LangDockerfile
	case "java":
		return LangJava
	case "kotlin", "kt":
		return LangKotlin
	case "rust", "rs":
		return LangRust
	case "ruby", "rb":
		return LangRuby
	case "lua":
		return LangLua
	case "yaml", "yml":
		return LangYaml
	case "sql":
		return LangSql
	case "css", "css3":
		return LangCss
	case "markdown", "md":
		return LangMarkdown
	case "sh", "bash":
		return LangBash
	case "csharp", "cs":
		return LangCsharp
	case "elixir", "ex":
		return LangElixir
	case "elm":
		return LangElm
	case "go":
		return LangGo
	case "groovy":
		return LangGroovy
	case "hcl", "tf":
		return LangHcl
	case "html":
		return LangHtml
	case "php":
		return LangPhp
	case "scala":
		return LangScala
	case "swift":
		return LangSwift
	default:
		return LangUnknown
	}
}

// ReadFromBytes reads a pattern checker definition from bytes array
func ReadFromBytes(fileContent []byte) (YamlChecker, error) {
	var checker PatternCheckerFile
	if err := yaml.Unmarshal(fileContent, &checker); err != nil {
		return nil, err
	}

	lang := DecodeLanguage(checker.Language)
	if lang == LangUnknown {
		return nil, fmt.Errorf("unknown language code: '%s'", checker.Language)
	}

	if checker.Code == "" {
		return nil, fmt.Errorf("no name provided in checker definition")
	}

	if checker.Message == "" {
		return nil, fmt.Errorf("no message provided in checker '%s'", checker.Code)
	}

	var patterns []*sitter.Query
	if checker.Pattern != "" {
		pattern, err := sitter.NewQuery([]byte(checker.Pattern), lang.Grammar())
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, pattern)
	} else if len(checker.Patterns) > 0 {
		for _, patternStr := range checker.Patterns {
			pattern, err := sitter.NewQuery([]byte(patternStr), lang.Grammar())
			if err != nil {
				return nil, err
			}
			patterns = append(patterns, pattern)
		}
	} else {
		return nil, fmt.Errorf("no pattern provided in checker '%s'", checker.Code)
	}

	if checker.Pattern != "" && len(checker.Patterns) > 0 {
		return nil, fmt.Errorf("only one of 'pattern' or 'patterns' can be provided in a checker definition")
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
				return nil, err
			}
			pathFilter.ExcludeGlobs = append(pathFilter.ExcludeGlobs, g)
		}

		for _, include := range checker.Include {
			g, err := glob.Compile(include)
			if err != nil {
				return nil, err
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
					return nil, err
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
					return nil, err
				}

				filters = append(filters, NodeFilter{
					query:       query,
					shouldMatch: false,
				})
			}
		}
	}

	patternChecker := &patternCheckerImpl{
		language:     lang,
		patterns:     patterns,
		issueMessage: checker.Message,
		issueId:      checker.Code,
		pathFilter:   pathFilter,
		filters:      filters,
	}

	return patternChecker, nil
}
