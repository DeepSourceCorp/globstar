package analysis

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/pkg/config"
)

type Issue struct {
	// The category of the issue
	Category config.Category
	// The severity of the issue
	Severity config.Severity
	// The message to display to the user
	Message string
	// The file path of the file that the issue was found in
	Filepath string
	// The range of the issue in the source code
	Range sitter.Range
	// (optional) The AST node that caused the issue
	Node *sitter.Node
	// Id is a unique ID for the issue.
	// Issue that have 'Id's can be explained using the `globstar desc` command.
	Id *string
}

func (i *Issue) AsJson() ([]byte, error) {
	type location struct {
		Row    int `json:"row"`
		Column int `json:"column"`
	}

	type position struct {
		Filename string   `json:"filename"`
		Start    location `json:"start"`
		End      location `json:"end"`
	}

	type issueJson struct {
		Category config.Category `json:"category"`
		Severity config.Severity `json:"severity"`
		Message  string          `json:"message"`
		Range    position        `json:"range"`
		Id       string          `json:"id"`
	}
	issue := issueJson{
		Category: i.Category,
		Severity: i.Severity,
		Message:  i.Message,
		Range: position{
			Filename: i.Filepath,
			Start: location{
				Row:    int(i.Range.StartPoint.Row),
				Column: int(i.Range.StartPoint.Column),
			},
			End: location{
				Row:    int(i.Range.EndPoint.Row),
				Column: int(i.Range.EndPoint.Column),
			},
		},
		Id: *i.Id,
	}

	return json.Marshal(issue)
}

func (i *Issue) AsText() ([]byte, error) {
	return []byte(fmt.Sprintf("%s:%d:%d:%s", i.Filepath, i.Range.StartPoint.Row, i.Range.StartPoint.Column, i.Message)), nil
}

type Analyzer struct {
	Language Language
	// WorkDir is the directory in which the analysis is being run.
	WorkDir string
	// ParseResult is the result of parsing a file with a tree-sitter parser,
	// along with some extra appendages (e.g: scope information).
	ParseResult *ParseResult
	// rules is a list of all rules that should be applied to the AST
	// for this language.
	rules []Rule
	// patternRules is a list of all rules that run after a query is run on the AST.
	// Usually, these are written in a DSL (which, for now, is the tree-sitter S-Expression query language)
	YmlRules []YmlRule
	// entryRules maps node types to the rules that should be applied
	// when entering that node.
	entryRulesForNode map[string][]Rule
	// exitRules maps node types to the rules that should be applied
	// when leaving that node.
	exitRulesForNode map[string][]Rule
	issuesRaised     []*Issue
}

func FromFile(filePath string, baseRules []Rule) (*Analyzer, error) {
	res, err := ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	return NewAnalyzer(res, baseRules), nil
}

func NewAnalyzer(file *ParseResult, rules []Rule) *Analyzer {
	ana := &Analyzer{
		ParseResult:       file,
		Language:          file.Language,
		entryRulesForNode: map[string][]Rule{},
		exitRulesForNode:  map[string][]Rule{},
	}

	for _, rule := range rules {
		ana.AddRule(rule)
	}

	return ana
}

func (ana *Analyzer) Analyze() []*Issue {
	WalkTree(ana.ParseResult.Ast, ana)
	ana.runPatternRules()
	return ana.issuesRaised
}

func (ana *Analyzer) AddRule(rule Rule) {
	ana.rules = append(ana.rules, rule)
	typ := rule.NodeType()

	if rule.OnEnter() != nil {
		ana.entryRulesForNode[typ] = append(ana.entryRulesForNode[typ], rule)
	}

	if rule.OnLeave() != nil {
		ana.exitRulesForNode[typ] = append(ana.exitRulesForNode[typ], rule)
	}
}

func (ana *Analyzer) OnEnterNode(node *sitter.Node) bool {
	nodeType := node.Type()
	rules := ana.entryRulesForNode[nodeType]
	for _, rule := range rules {
		visitFn := rule.OnEnter()
		if visitFn != nil {
			(*visitFn)(rule, ana, node)
		}
	}
	return true
}

func (ana *Analyzer) OnLeaveNode(node *sitter.Node) {
	nodeType := node.Type()
	rules := ana.exitRulesForNode[nodeType]
	for _, rule := range rules {
		visitFn := rule.OnLeave()
		if visitFn != nil {
			(*visitFn)(rule, ana, node)
		}
	}
}

func (ana *Analyzer) shouldSkipRule(rule YmlRule) bool {
	pathFilter := rule.PathFilter()
	if pathFilter == nil {
		// no filter is set, so we should not skip this rule
		return false
	}

	relPath := ana.ParseResult.FilePath
	if ana.WorkDir != "" {
		rel, err := filepath.Rel(ana.WorkDir, ana.ParseResult.FilePath)
		if err == nil {
			relPath = rel
		}
	}

	if len(pathFilter.ExcludeGlobs) > 0 {
		for _, excludeGlob := range pathFilter.ExcludeGlobs {
			if excludeGlob.Match(relPath) {
				return true
			}
		}

		// no exclude globs matched, so we should not skip this rule
		return false
	}

	if len(pathFilter.IncludeGlobs) > 0 {
		for _, includeGlob := range pathFilter.IncludeGlobs {
			if includeGlob.Match(relPath) {
				return false
			}
		}

		// no include globs matched, so we should skip this rule
		return true
	}

	return false
}

func (ana *Analyzer) filterMatchesParent(filter *NodeFilter, parent *sitter.Node) bool {
	qc := sitter.NewQueryCursor()
	defer qc.Close()

	qc.Exec(filter.query, parent)

	// check if the filter matches the `parent` node
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}

		m = qc.FilterPredicates(m, ana.ParseResult.Source)
		for _, capture := range m.Captures {
			captureName := filter.query.CaptureNameForId(capture.Index)
			if captureName == filterPatternKey && capture.Node == parent {
				return true
			}
		}
	}

	return false
}

// runParentFilters checks if the parent filters for a rule match the given node.
func (ana *Analyzer) runParentFilters(rule YmlRule, node *sitter.Node) bool {
	filters := rule.NodeFilters()
	if len(filters) == 0 {
		return true
	}

	for _, filter := range filters {
		shouldMatch := filter.shouldMatch
		nodeMatched := false

		// The matched node is expected to be a child of some other
		// node, but it has no parents (is a top-level node)
		if node.Parent() == nil && filter.shouldMatch {
			return false
		}

		for parent := node.Parent(); parent != nil; parent = parent.Parent() {
			if ana.filterMatchesParent(&filter, parent) {
				nodeMatched = true
				if !shouldMatch {
					// pattern-not-inside matched, so this rule should be skipped
					return false
				} else {
					// pattern-inside matched, so we can break out of the loop
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

func (ana *Analyzer) executeRuleQuery(rule YmlRule, query *sitter.Query) {
	qc := sitter.NewQueryCursor()
	defer qc.Close()

	qc.Exec(query, ana.ParseResult.Ast)
	for {
		m, ok := qc.NextMatch()

		if !ok {
			break
		}

		m = qc.FilterPredicates(m, ana.ParseResult.Source)
		for _, capture := range m.Captures {
			captureName := query.CaptureNameForId(capture.Index)
			// TODO: explain why captureName == rule.Name()
			if captureName == rule.Name() && ana.runParentFilters(rule, capture.Node) {
				rule.OnMatch(ana, query, capture.Node, m.Captures)
			}
		}
	}
}

// runPatternRules executes all rules that are written as AST queries.
func (ana *Analyzer) runPatternRules() {
	for _, rule := range ana.YmlRules {
		if ana.shouldSkipRule(rule) {
			continue
		}

		queries := rule.Patterns()
		for _, q := range queries {
			ana.executeRuleQuery(rule, q)
		}
	}
}

func (ana *Analyzer) Report(issue *Issue) {
	ana.issuesRaised = append(ana.issuesRaised, issue)
}
