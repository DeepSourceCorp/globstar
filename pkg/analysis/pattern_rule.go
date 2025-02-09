package analysis

import (
	"fmt"
	"os"
	"strings"

	"github.com/DeepSourceCorp/globstar/pkg/config"
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
// We need a to append a key name at the end of the pattern written by the user.
// This is the key that we will use.
const filterPatternKey = "__filter__key__"

// A YmlRule is a rule that matches a tree-sitter query pattern
// and reports an issue when the pattern is found.
// Unlike regular issues, PatternRules are not associated with a specific node type, rather
// they are invoked for *every* node that matches the pattern.
type YmlRule interface {
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

// NodeFilter is a filter that can be applied to a PatternRule to restrict
// the the nodes that the rule is applied to.
// The rule is only applied to nodes that have a parent matching (or not matching) the query.
type NodeFilter struct {
	query       *sitter.Query
	shouldMatch bool
}

// PathFilter is a glob that can be applied to a PatternRule to restrict
// the files that the rule is applied to.
type PathFilter struct {
	ExcludeGlobs []glob.Glob
	IncludeGlobs []glob.Glob
}

type patternRuleImpl struct {
	language     Language
	patterns     []*sitter.Query
	issueMessage string
	issueId      string
	category     config.Category
	severity     config.Severity
	pathFilter   *PathFilter
	filters      []NodeFilter
}

func (r *patternRuleImpl) Language() Language {
	return r.language
}

func (r *patternRuleImpl) Patterns() []*sitter.Query {
	return r.patterns
}

func (r *patternRuleImpl) OnMatch(
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

	ana.Report(&Issue{
		Range:    matchedNode.Range(),
		Message:  message,
		Category: r.Category(),
		Severity: r.Severity(),
		Id:       &r.issueId,
	})
}

func (r *patternRuleImpl) Name() string {
	return r.issueId
}

func (r *patternRuleImpl) PathFilter() *PathFilter {
	return r.pathFilter
}

func (r *patternRuleImpl) NodeFilters() []NodeFilter {
	return r.filters
}

func (r *patternRuleImpl) Category() config.Category {
	return r.category
}

func (r *patternRuleImpl) Severity() config.Severity {
	return r.severity
}

func CreatePatternRule(
	patterns []*sitter.Query,
	language Language,
	issueMessage string,
	issueId string,
	pathFilter *PathFilter,
) YmlRule {
	return &patternRuleImpl{
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

type PatternRuleFile struct {
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

// ReadFromFile reads a pattern rule definition from a YAML config file.
func ReadFromFile(filePath string) (YmlRule, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return ReadFromBytes(fileContent)
}

// ReadFromBytes reads a pattern rule definition from bytes array
func ReadFromBytes(fileContent []byte) (YmlRule, error) {
	var rule PatternRuleFile
	if err := yaml.Unmarshal(fileContent, &rule); err != nil {
		return nil, err
	}

	lang := DecodeLanguage(rule.Language)
	if lang == LangUnknown {
		return nil, fmt.Errorf("unknown language code: '%s'", rule.Language)
	}

	var patterns []*sitter.Query
	if rule.Pattern != "" {
		pattern, err := sitter.NewQuery([]byte(rule.Pattern), lang.Grammar())
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, pattern)
	} else if len(rule.Patterns) > 0 {
		for _, patternStr := range rule.Patterns {
			pattern, err := sitter.NewQuery([]byte(patternStr), lang.Grammar())
			if err != nil {
				return nil, err
			}
			patterns = append(patterns, pattern)
		}
	} else {
		return nil, fmt.Errorf("no pattern provided in rule '%s'", rule.Code)
	}

	if rule.Pattern != "" && len(rule.Patterns) > 0 {
		return nil, fmt.Errorf("only one of 'pattern' or 'patterns' can be provided in a rule definition")
	}

	// include and exclude patterns
	var pathFilter *PathFilter
	if rule.Exclude != nil || rule.Include != nil {
		pathFilter = &PathFilter{
			ExcludeGlobs: make([]glob.Glob, 0, len(rule.Exclude)),
			IncludeGlobs: make([]glob.Glob, 0, len(rule.Include)),
		}

		for _, exclude := range rule.Exclude {
			g, err := glob.Compile(exclude)
			if err != nil {
				return nil, err
			}
			pathFilter.ExcludeGlobs = append(pathFilter.ExcludeGlobs, g)
		}

		for _, include := range rule.Include {
			g, err := glob.Compile(include)
			if err != nil {
				return nil, err
			}
			pathFilter.IncludeGlobs = append(pathFilter.IncludeGlobs, g)
		}
	}

	// node filters
	var filters []NodeFilter
	if rule.Filters != nil {
		for _, filter := range rule.Filters {
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

	patternRule := &patternRuleImpl{
		language:     lang,
		patterns:     patterns,
		issueMessage: rule.Message,
		issueId:      rule.Code,
		pathFilter:   pathFilter,
		filters:      filters,
	}

	return patternRule, nil
}
