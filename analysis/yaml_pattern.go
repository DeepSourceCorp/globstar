package analysis

import (
	"github.com/gobwas/glob"
	sitter "github.com/smacker/go-tree-sitter"
	"strings"
)

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

type YamlChecker interface {
	Name() string
	Patterns() []*sitter.Query
	Language() Language
	Category() Category
	Severity() Severity
	OnMatch(
		ana *YamlAnalyzer, // the analyzer instance
		matchedQuery *sitter.Query, // the query that found an AST node
		matchedNode *sitter.Node, // the AST node that matched the query
		captures []sitter.QueryCapture, // list of captures made inside the query
	)
	PathFilter() *PathFilter
	NodeFilters() []NodeFilter
}

type patternCheckerImpl struct {
	language     Language
	patterns     []*sitter.Query
	issueMessage string
	issueId      string
	category     Category
	severity     Severity
	pathFilter   *PathFilter
	filters      []NodeFilter
}

var fileSkipComment = make(map[string][]*SkipComment)

func (r *patternCheckerImpl) Language() Language {
	return r.language
}

func (r *patternCheckerImpl) Patterns() []*sitter.Query {
	return r.patterns
}

func (r *patternCheckerImpl) OnMatch(
	ana *YamlAnalyzer,
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
	raisedIssue := &YamlIssue{
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
	if ContainsSkipcqYaml(skipComments, raisedIssue) {
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

func (r *patternCheckerImpl) Category() Category {
	return r.category
}

func (r *patternCheckerImpl) Severity() Severity {
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
	Language string   `yaml:"language"`
	Code     string   `yaml:"name"`
	Message  string   `yaml:"message"`
	Category Category `yaml:"category"`
	Severity Severity `yaml:"severity"`
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
