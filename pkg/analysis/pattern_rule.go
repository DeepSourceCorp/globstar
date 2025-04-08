package analysis

import (
	"fmt"
	"os"
	"strings"

	"github.com/gobwas/glob"
	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/pkg/config"
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
// We need a to append a key name at the end of the pattern written by the user.
// This is the key that we will use.
const filterPatternKey = "__filter__key__"

// A YamlChecker is a checker that matches a tree-sitter query pattern
// and reports an issue when the pattern is found.
// Unlike regular issues, PatternCheckers are not associated with a specific node type, rather
// they are invoked for *every* node that matches the pattern.
type YamlChecker interface {
	Name() string
	Patterns() []*sitter.Query
	Language() Language
	Category() config.Category
	Severity() config.Severity
	OnMatch(
		ana *Analyzer, // the analyzer instance
		matchedQuery *sitter.Query, // the query that found an AST node
		matchedNode *sitter.Node, // the AST node that matched the query
		captures []sitter.QueryCapture, // list of captures made inside the query
	)
	PathFilter() *PathFilter
	NodeFilters() []NodeFilter
}

// NodeFilter is a filter that can be applied to a PatternChecker to restrict
// the the nodes that the checker is applied to.
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

type patternCheckerImpl struct {
	language     Language
	patterns     []*sitter.Query
	issueMessage string
	issueId      string
	category     config.Category
	severity     config.Severity
	pathFilter   *PathFilter
	filters      []NodeFilter
}

func (r *patternCheckerImpl) Language() Language {
	return r.language
}

func (r *patternCheckerImpl) Patterns() []*sitter.Query {
	return r.patterns
}

func (r *patternCheckerImpl) OnMatch(
	ana *Analyzer,
	matchedQuery *sitter.Query,
	matchedNode *sitter.Node,
	captures []sitter.QueryCapture,
) {

	// replace all '@<capture-name>' with the corresponding capture value
	message := r.issueMessage
	// TODO: 1. escape '@' in the message, 2. use a more efficient way to replace
	for strings.ContainsRune(message, '@') {
		for _, capture := range captures {
			captureName := matchedQuery.CaptureNameForId(capture.Index)
			message = strings.ReplaceAll(
				message,
				"@"+captureName,
				capture.Node.Content(ana.ParseResult.Source),
			)
		}
	}
	raisedIssue := &Issue{
		Range:    matchedNode.Range(),
		Node:     matchedNode,
		Message:  message,
		Filepath: ana.ParseResult.FilePath,
		Category: r.Category(),
		Severity: r.Severity(),
		Id:       &r.issueId,
	}

	filepath := ana.ParseResult.FilePath
	skipComments := fileSkipComment[filepath]
	if !ana.ContainsSkipcq(skipComments, raisedIssue) {
		ana.Report(raisedIssue)
	}
}

func (r *patternCheckerImpl) Name() string {
	return r.issueId
}

func (r *patternCheckerImpl) PathFilter() *PathFilter {
	return r.pathFilter
}

func (r *patternCheckerImpl) NodeFilters() []NodeFilter {
	return r.filters
}

func (r *patternCheckerImpl) Category() config.Category {
	return r.category
}

func (r *patternCheckerImpl) Severity() config.Severity {
	return r.severity
}

func CreatePatternChecker(
	patterns []*sitter.Query,
	language Language,
	issueMessage string,
	issueId string,
	pathFilter *PathFilter,
) YamlChecker {
	return &patternCheckerImpl{
		language:     language,
		patterns:     patterns,
		issueMessage: issueMessage,
		issueId:      issueId,
		pathFilter:   pathFilter,
	}
}

type filterYAML struct {
	PatternInside    string `yaml:"pattern-inside,omitempty"`
	PatternNotInside string `yaml:"pattern-not-inside,omitempty"`
}

type PatternCheckerFile struct {
	Language string          `yaml:"language"`
	Code     string          `yaml:"name"`
	Message  string          `yaml:"message"`
	Category config.Category `yaml:"category"`
	Severity config.Severity `yaml:"severity"`
	// Pattern is a single pattern in the form of:
	// pattern: (some_pattern)
	// in the YAML file
	Pattern string `yaml:"pattern,omitempty"`
	// Patterns are ultiple patterns in the form of:
	// pattern:  (something)
	// in the YAML file
	Patterns    []string     `yaml:"patterns,omitempty"`
	Description string       `yaml:"description,omitempty"`
	Filters     []filterYAML `yaml:"filters,omitempty"`
	Exclude     []string     `yaml:"exclude,omitempty"`
	Include     []string     `yaml:"include,omitempty"`
}

// DecodeLanguage converts a stringified language name to its corresponding
// Language enum
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

// ReadFromFile reads a pattern checker definition from a YAML config file.
func ReadFromFile(filePath string) (YamlChecker, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return ReadFromBytes(fileContent)
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
