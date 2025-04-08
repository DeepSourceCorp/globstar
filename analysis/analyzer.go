package analysis

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"

	sitter "github.com/smacker/go-tree-sitter"
)

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityError    Severity = "error"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
)

func (s Severity) IsValid() bool {
	switch s {
	case SeverityCritical, SeverityError, SeverityWarning, SeverityInfo:
		return true
	}
	return false
}

type Category string

const (
	CategoryStyle       Category = "style"
	CategoryBugRisk     Category = "bug-risk"
	CategoryAntipattern Category = "antipattern"
	CategoryPerformance Category = "performance"
	CategorySecurity    Category = "security"
)

func (c Category) IsValid() bool {
	switch c {
	case CategoryStyle, CategoryBugRisk, CategoryAntipattern, CategoryPerformance, CategorySecurity:
		return true
	}
	return false
}

type Analyzer struct {
	Name        string
	Description string
	Category    Category
	Severity    Severity
	Language    Language
	Requires    []*Analyzer
	Run         func(*Pass) (any, error)
	ResultType  reflect.Type
}

type Pass struct {
	Analyzer    *Analyzer
	FileContext *ParseResult
	Files       []*ParseResult
	ResultOf    map[*Analyzer]any
	Report      func(*Pass, *sitter.Node, string)
	// TODO (opt): the cache should ideally not be stored in-memory
	ResultCache map[*Analyzer]map[*ParseResult]any
}

// for caching the skipcq comments
type SkipComment struct {
	// the line number for the skipcq comment
	CommentLine int
	// the entire text of the skipcq comment
	CommentText string
	// (optional) name of the checker for targetted skip
	CheckerId string
}

func walkTree(node *sitter.Node, f func(*sitter.Node)) {
	f(node)

	for i := 0; i < int(node.ChildCount()); i++ {
		walkTree(node.Child(i), f)
	}
}

func Preorder(pass *Pass, fn func(*sitter.Node)) {
	// TODO: cache the traversal results to avoid running the traversal for each analyzer
	walkTree(pass.FileContext.Ast, fn)
}

var defaultIgnoreDirs = []string{
	"checkers",
	"node_modules",
	"vendor",
	"dist",
	"build",
	"out",
	".git",
	".svn",
	"venv",
	"__pycache__",
	".idea",
	".vitepress",
}

func findAnalyzers(analyzer *Analyzer) []*Analyzer {
	analyzers := []*Analyzer{}
	for _, req := range analyzer.Requires {
		analyzers = append(analyzers, findAnalyzers(req)...)
	}
	analyzers = append(analyzers, analyzer)
	return analyzers
}

func RunAnalyzers(path string, analyzers []*Analyzer, fileFilter func(string) bool) ([]*Issue, error) {
	raisedIssues := []*Issue{}
	langAnalyzerMap := make(map[Language][]*Analyzer)

	for _, analyzer := range analyzers {
		langAnalyzerMap[analyzer.Language] = append(langAnalyzerMap[analyzer.Language], findAnalyzers(analyzer)...)
	}

	trees := make(map[Language][]*ParseResult)
	fileSkipInfo := make(map[string][]*SkipComment)

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // continue to the next file
		}

		if info.IsDir() {
			if slices.Contains(defaultIgnoreDirs, info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if fileFilter != nil && !fileFilter(path) {
			return nil
		}

		file, err := ParseFile(path)
		if err != nil {
			if err != ErrUnsupportedLanguage {
				fmt.Println(err)
			}
			return nil
		}

		fileSkipInfo[file.FilePath] = GatherSkipInfo(file)

		trees[file.Language] = append(trees[file.Language], file)

		return nil
	})
	if err != nil {
		return raisedIssues, err
	}

	reportFunc := func(pass *Pass, node *sitter.Node, message string) {
		raisedIssue := &Issue{
			Id:       &pass.Analyzer.Name,
			Node:     node,
			Message:  message,
			Filepath: pass.FileContext.FilePath,
		}

		skipLines := fileSkipInfo[pass.FileContext.FilePath]
		if !ContainsSkipcq(skipLines, raisedIssue) {
			raisedIssues = append(raisedIssues, raisedIssue)
		}
	}

	for lang, analyzers := range langAnalyzerMap {
		pass := &Pass{
			Files:       trees[lang],
			Report:      reportFunc,
			ResultOf:    make(map[*Analyzer]any),
			ResultCache: make(map[*Analyzer]map[*ParseResult]any),
		}

		for _, file := range pass.Files {
			pass.FileContext = file
			for _, analyzer := range analyzers {
				pass.Analyzer = analyzer

				if len(pass.Files) == 0 {
					continue
				}

				result, err := analyzer.Run(pass)
				if err != nil {
					return raisedIssues, err
				}

				pass.ResultOf[analyzer] = result
				if _, ok := pass.ResultCache[analyzer]; !ok {
					pass.ResultCache[analyzer] = make(map[*ParseResult]any)
				}
				pass.ResultCache[analyzer][file] = result
			}
		}
	}

	return raisedIssues, nil
}

func ReportIssues(issues []*Issue, format string) ([]byte, error) {
	switch format {
	case "json":
		return reportJSON(issues)
	case "text":
		return reportText(issues)
	default:
		return reportText(issues)
	}
}

func reportJSON(issues []*Issue) ([]byte, error) {
	output := []byte{}
	for _, issue := range issues {
		json, err := issue.AsJson()
		if err != nil {
			return []byte{}, err
		}
		output = append(output, json...)
		output = append(output, []byte("\n")...)
	}
	return output, nil
}

func reportText(issues []*Issue) ([]byte, error) {
	output := []byte{}
	for _, issue := range issues {
		issueText, err := issue.AsText()
		if err != nil {
			return []byte{}, err
		}
		output = append(output, issueText...)
		output = append(output, []byte("\n")...)
	}
	return output, nil
}

// cache all the skipcq comments from an ast
func GatherSkipInfo(fileContext *ParseResult) []*SkipComment {
	var skipLines []*SkipComment

	commentIdentifier := GetEscapedCommentIdentifierFromPath(fileContext.FilePath)
	pattern := fmt.Sprintf(`^%s\s+skipcq(?::\s*([A-Za-z0-9_-]+))?`, commentIdentifier)
	skipRegexp := regexp.MustCompile(pattern)

	query, err := sitter.NewQuery([]byte("(comment) @skipcq"), fileContext.Language.
		Grammar())

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

			// look for checker names
			matches := skipRegexp.FindStringSubmatch(commentText)
			var checkerId string
			if matches != nil {
				if len(matches) > 1 {
					checkerId = matches[1]
				}
			}

			if skipRegexp.MatchString(commentText) {
				skipLines = append(skipLines, &SkipComment{
					CommentLine: commentLine,
					CommentText: commentText,
					CheckerId:   checkerId, // will be empty for generic skipcq
				})
			}
		}
	}

	return skipLines
}

func ContainsSkipcq(skipLines []*SkipComment, issue *Issue) bool {
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

		if comment.CheckerId != "" {
			// targetted skipcq
			if checkerId == comment.CheckerId {
				return true
			}
		} else {
			// generic skipcq
			return true
		}
	}

	return false
}
