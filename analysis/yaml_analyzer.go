package analysis

import (
	sitter "github.com/smacker/go-tree-sitter"
	"fmt"
	"regexp"
	"strings"
	"encoding/json"
	"path/filepath"
)

type YamlIssue struct {
	// The category of the issue
	Category Category
	// The severity of the issue
	Severity Severity
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

func (i *YamlIssue) AsJson() ([]byte, error) {
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
		Category Category `json:"category"`
		Severity Severity `json:"severity"`
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

func (i *YamlIssue) AsText() ([]byte, error) {
	return []byte(fmt.Sprintf("%s:%d:%d:%s", i.Filepath, i.Range.StartPoint.Row, i.Range.StartPoint.Column, i.Message)), nil
}

type YamlAnalyzer struct {
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
	issuesRaised        []*YamlIssue
}

func InitializeSkipComments(analyzers []*YamlAnalyzer) {
	fileSkipComments := make(map[string][]*SkipComment)

	processedPaths := make(map[string]bool)

	for _, analyzer := range analyzers {
		filepath := analyzer.ParseResult.FilePath
		if processedPaths[filepath] {
			continue
		}

		processedPaths[filepath] = true
		fileSkipComments[filepath] = GatherSkipInfo(analyzer.ParseResult)
	}
}

func FromFile(filePath string, baseCheckers []Checker) (*YamlAnalyzer, error) {
	res, err := ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	return NewAnalyzer(res, baseCheckers), nil
}

func NewAnalyzer(file *ParseResult, checkers []Checker) *YamlAnalyzer {
	ana := &YamlAnalyzer{
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

func (ana *YamlAnalyzer) Analyze() []*YamlIssue {
	WalkTree(ana.ParseResult.Ast, ana)
	ana.runPatternCheckers()
	return ana.issuesRaised
}

func (ana *YamlAnalyzer) AddChecker(checker Checker) {
	ana.checkers = append(ana.checkers, checker)
	typ := checker.NodeType()

	if checker.OnEnter() != nil {
		ana.entryCheckersForNode[typ] = append(ana.entryCheckersForNode[typ], checker)
	}

	if checker.OnLeave() != nil {
		ana.exitCheckersForNode[typ] = append(ana.exitCheckersForNode[typ], checker)
	}
}

func (ana *YamlAnalyzer) OnEnterNode(node *sitter.Node) bool {
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

func (ana *YamlAnalyzer) OnLeaveNode(node *sitter.Node) {
	nodeType := node.Type()
	checkers := ana.exitCheckersForNode[nodeType]
	for _, checker := range checkers {
		visitFn := checker.OnLeave()
		if visitFn != nil {
			(*visitFn)(checker, ana, node)
		}
	}
}

func (ana *YamlAnalyzer) shouldSkipChecker(checker YamlChecker) bool {
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

func (ana *YamlAnalyzer) filterMatchesParent(filter *NodeFilter, parent *sitter.Node) bool {
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

func (ana *YamlAnalyzer) runParentFilters(checker YamlChecker, node *sitter.Node) bool {
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

func (ana *YamlAnalyzer) executeCheckerQuery(checker YamlChecker, query *sitter.Query) {
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
func (ana *YamlAnalyzer) runPatternCheckers() {
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

func (ana *YamlAnalyzer) Report(issue *YamlIssue) {
	ana.issuesRaised = append(ana.issuesRaised, issue)
}

func RunYamlCheckers(path string, analyzers []*YamlAnalyzer) ([]*YamlIssue, error) {
	InitializeSkipComments(analyzers)

	var issues []*YamlIssue
	for _, analyzer := range analyzers {
		issues = append(issues, analyzer.Analyze()...)
	}
	return issues, nil
}

func YamlGatherSkipInfo(fileContext *ParseResult) []*SkipComment {
	var skipLines []*SkipComment

	commentIdentifier := GetEscapedCommentIdentifierFromPath(fileContext.FilePath)
	pattern := fmt.Sprintf(`%s(?i).*?\bskipcq\b(?::(?:\s*(?P<issue_ids>([A-Za-z\-_0-9]*(?:,\s*)?)+))?)?`, commentIdentifier)
	skipRegexp := regexp.MustCompile(pattern)

	query, err := sitter.NewQuery([]byte("(comment) @skipcq"), fileContext.Language.Grammar())

	if err != nil {
		return skipLines
	}

	cursor := sitter.NewQueryCursor()
	cursor.Exec(query, fileContext.Ast)

	// gather all skipcq comment lines in a single pass
	for {
		m, ok := cursor.NextMatch()
		if !ok {
			break
		}

		for _, capture := range m.Captures {
			captureName := query.CaptureNameForId(capture.Index)
			if captureName != "skipcq" {
				continue
			}

			commentNode := capture.Node
			commentLine := int(commentNode.StartPoint().Row)
			commentText := commentNode.Content(fileContext.Source)

			matches := skipRegexp.FindStringSubmatch(commentText)
			if matches != nil {
				issueIdsIdx := skipRegexp.SubexpIndex("issue_ids")
				var checkerIds []string

				if issueIdsIdx != -1 && issueIdsIdx < len(matches) && matches[issueIdsIdx] != "" {
					issueIdsIdx := matches[issueIdsIdx]
					idSlice := strings.Split(issueIdsIdx, ",")
					for _, id := range idSlice {
						trimmedId := strings.TrimSpace(id)
						if trimmedId != "" {
							checkerIds = append(checkerIds, trimmedId)
						}
					}
				}

				skipLines = append(skipLines, &SkipComment{
					CommentLine: commentLine,
					CommentText: commentText,
					CheckerIds:  checkerIds, // will be empty for generic skipcq
				})
			}

		}
	}

	return skipLines
}

func ContainsSkipcqYaml(skipLines []*SkipComment, issue *YamlIssue) bool {
	if len(skipLines) == 0 {
		return false
	}

	issueNode := issue.Node
	nodeLine := int(issueNode.StartPoint().Row)
	prevLine := nodeLine - 1

	var checkerId string
	if issue.Id != nil {
		checkerId = *issue.Id
	}

	for _, comment := range skipLines {
		if comment.CommentLine != nodeLine && comment.CommentLine != prevLine {
			continue
		}

		if len(comment.CheckerIds) > 0 {
			for _, id := range comment.CheckerIds {
				if checkerId == id {
					return true
				}
			}
		} else {
			return true
		}
	}

	return false
}

