package one

import (
	"fmt"
	"os"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"gopkg.in/yaml.v3"
)

// A PatternRule is a rule that matches a tree-sitter query pattern
// and reports an issue when the pattern is found.
// Unlike regular issues, PatternRules are not associated with a specific node type, rather
// they are invoked for *every* node that matches the pattern.
type PatternRule interface {
	Language() Language
	Pattern() *sitter.Query
	OnMatch(ana *Analyzer, matchedNode *sitter.Node)
}

type patternRuleImpl struct {
	language     Language
	pattern      *sitter.Query
	issueMessage string
	issueId      *string
}

func (r *patternRuleImpl) Language() Language {
	return r.language
}

func (r *patternRuleImpl) Pattern() *sitter.Query {
	return r.pattern
}

func (r *patternRuleImpl) OnMatch(ana *Analyzer, matchedNode *sitter.Node) {
	ana.Report(&Issue{
		Range:   matchedNode.Range(),
		Message: r.issueMessage,
		Id:      r.issueId,
	})
}

func CreatePatternRule(pattern *sitter.Query, language Language, issueMessage string, issueId *string) PatternRule {
	return &patternRuleImpl{
		language:     language,
		pattern:      pattern,
		issueMessage: issueMessage,
		issueId:      issueId,
	}
}

type PatternRuleFile struct {
	Language    string `yaml:"language"`
	Code        string `yaml:"Code"`
	Message     string `yaml:"message"`
	Pattern     string `yaml:"pattern"`
	Description string `yaml:"description"`
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

	patternRule := CreatePatternRule(pattern, lang, rule.Message, &rule.Code)
	return patternRule, nil
}
