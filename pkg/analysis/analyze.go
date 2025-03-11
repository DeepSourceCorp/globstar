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
	// checkers is a list of all checkers that should be applied to the AST
	// for this language.
	checkers []Checker
	// patternCheckers is a list of all checkers that run after a query is run on the AST.
	// Usually, these are written in a DSL (which, for now, is the tree-sitter S-Expression query language)
	YamlCheckers []YamlChecker
	// entryCheckers maps node types to the checkers that should be applied
	// when entering that node.
	entryCheckersForNode map[string][]Checker
	// exitCheckers maps node types to the checkers that should be applied
	// when leaving that node.
	exitCheckersForNode map[string][]Checker
	issuesRaised        []*Issue
}

func FromFile(filePath string, baseCheckers []Checker) (*Analyzer, error) {
	res, err := ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	return NewAnalyzer(res, baseCheckers), nil
}

func NewAnalyzer(file *ParseResult, checkers []Checker) *Analyzer {
	ana := &Analyzer{
		ParseResult:          file,
		Language:             file.Language,
		entryCheckersForNode: map[string][]Checker{},
		exitCheckersForNode:  map[string][]Checker{},
	}

	for _, checker := range checkers {
		ana.AddChecker(checker)
	}

	return ana
}

func (ana *Analyzer) Analyze() []*Issue {
	WalkTree(ana.ParseResult.Ast, ana)
	ana.runPatternCheckers()
	return ana.issuesRaised
}

func (ana *Analyzer) AddChecker(checker Checker) {
	ana.checkers = append(ana.checkers, checker)
	typ := checker.NodeType()

	if checker.OnEnter() != nil {
		ana.entryCheckersForNode[typ] = append(ana.entryCheckersForNode[typ], checker)
	}

	if checker.OnLeave() != nil {
		ana.exitCheckersForNode[typ] = append(ana.exitCheckersForNode[typ], checker)
	}
}

func (ana *Analyzer) OnEnterNode(node *sitter.Node) bool {
	nodeType := node.Type()
	checkers := ana.entryCheckersForNode[nodeType]
	for _, checker := range checkers {
		visitFn := checker.OnEnter()
		if visitFn != nil {
			(*visitFn)(checker, ana, node)
		}
	}
	return true
}

func (ana *Analyzer) OnLeaveNode(node *sitter.Node) {
	nodeType := node.Type()
	checkers := ana.exitCheckersForNode[nodeType]
	for _, checker := range checkers {
		visitFn := checker.OnLeave()
		if visitFn != nil {
			(*visitFn)(checker, ana, node)
		}
	}
}

func (ana *Analyzer) shouldSkipChecker(checker YamlChecker) bool {
	pathFilter := checker.PathFilter()
	if pathFilter == nil {
		// no filter is set, so we should not skip this checker
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

		// no exclude globs matched, so we should not skip this checker
		return false
	}

	if len(pathFilter.IncludeGlobs) > 0 {
		for _, includeGlob := range pathFilter.IncludeGlobs {
			if includeGlob.Match(relPath) {
				return false
			}
		}

		// no include globs matched, so we should skip this checker
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

// runParentFilters checks if the parent filters for a checker match the given node.
func (ana *Analyzer) runParentFilters(checker YamlChecker, node *sitter.Node) bool {
	filters := checker.NodeFilters()
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
					// pattern-not-inside matched, so this checker should be skipped
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

func (ana *Analyzer) executeCheckerQuery(checker YamlChecker, query *sitter.Query) {
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
			// TODO: explain why captureName == checker.Name()
			if captureName == checker.Name() && ana.runParentFilters(checker, capture.Node) {
				checker.OnMatch(ana, query, capture.Node, m.Captures)
			}
		}
	}
}

// runPatternCheckers executes all checkers that are written as AST queries.
func (ana *Analyzer) runPatternCheckers() {
	for _, checker := range ana.YamlCheckers {
		if ana.shouldSkipChecker(checker) {
			continue
		}

		queries := checker.Patterns()
		for _, q := range queries {
			ana.executeCheckerQuery(checker, q)
		}
	}
}

func (ana *Analyzer) Report(issue *Issue) {
	ana.issuesRaised = append(ana.issuesRaised, issue)
}

func RunYamlCheckers(path string, analyzers []*Analyzer) ([]*Issue, error) {
	issues := []*Issue{}
	for _, analyzer := range analyzers {
		issues = append(issues, analyzer.Analyze()...)
	}
	return issues, nil
}
