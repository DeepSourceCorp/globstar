package one

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
// We need a to append a key name at the end of the pattern written by the user.
// This is the key that we will use.
const filterPatternKey = "__filter__key__"

// A PatternRule is a rule that matches a tree-sitter query pattern
// and reports an issue when the pattern is found.
// Unlike regular issues, PatternRules are not associated with a specific node type, rather
// they are invoked for *every* node that matches the pattern.
type PatternRule interface {
	Name() string
	Pattern() *sitter.Query
	Language() Language
	OnMatch(ana *Analyzer, matchedNode *sitter.Node, captures []sitter.QueryCapture)
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
	pattern      *sitter.Query
	issueMessage string
	issueId      string
	pathFilter   *PathFilter
	filters      []NodeFilter
}

func (r *patternRuleImpl) Language() Language {
	return r.language
}

func (r *patternRuleImpl) Pattern() *sitter.Query {
	return r.pattern
}

func (r *patternRuleImpl) OnMatch(
	ana *Analyzer,
	matchedNode *sitter.Node,
	captures []sitter.QueryCapture,
) {

	// replace all '@<capture-name>' with the corresponding capture value
	message := r.issueMessage
	// TODO: 1. escape '@' in the message, 2. use a more efficient way to replace
	for strings.ContainsRune(message, '@') {
		for _, capture := range captures {
			captureName := r.pattern.CaptureNameForId(capture.Index)
			message = strings.ReplaceAll(
				message,
				"@"+captureName,
				capture.Node.Content(ana.ParseResult.Source),
			)
		}
	}

	ana.Report(&Issue{
		Range:   matchedNode.Range(),
		Message: message,
		Id:      &r.issueId,
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

func CreatePatternRule(pattern *sitter.Query,
	language Language,
	issueMessage string,
	issueId string,
	pathFilter *PathFilter,
) PatternRule {
	return &patternRuleImpl{
		language:     language,
		pattern:      pattern,
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
	Language    string       `yaml:"language"`
	Code        string       `yaml:"name"`
	Message     string       `yaml:"message"`
	Pattern     string       `yaml:"pattern"`
	Description string       `yaml:"description,omitempty"`
	Filters     []filterYAML `yaml:"filters,omitempty"`
	Exclude     []string     `yaml:"exclude,omitempty"`
	Include     []string     `yaml:"include,omitempty"`
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
	default:
		return LangUnknown
	}
}

// ReadFromFile reads a pattern rule definition from a YAML config file.
func ReadFromFile(filePath string) (PatternRule, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var rule PatternRuleFile
	if err = yaml.Unmarshal(fileContent, &rule); err != nil {
		return nil, err
	}

	lang := DecodeLanguage(rule.Language)
	if lang == LangUnknown {
		return nil, fmt.Errorf("unknown language code: '%s'", rule.Language)
	}

	pattern, err := sitter.NewQuery([]byte(rule.Pattern), lang.Grammar())
	if err != nil {
		return nil, err
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
		pattern:      pattern,
		issueMessage: rule.Message,
		issueId:      rule.Code,
		pathFilter:   pathFilter,
		filters:      filters,
	}

	return patternRule, nil
}
