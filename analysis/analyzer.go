package analysis

import (
	"fmt"
	"os"
	"path/filepath"
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
	Run         func(*Pass) (interface{}, error)
}

type Pass struct {
	Analyzer    *Analyzer
	FileContext *ParseResult
	Files       []*ParseResult
	Report      func(*Pass, *sitter.Node, string)
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

func RunAnalyzers(path string, analyzers []*Analyzer, fileFilter func(string) bool) ([]*Issue, error) {
	raisedIssues := []*Issue{}
	langAnalyzerMap := make(map[Language][]*Analyzer)
	for _, analyzer := range analyzers {
		langAnalyzerMap[analyzer.Language] = append(langAnalyzerMap[analyzer.Language], analyzer)
	}

	trees := make(map[Language][]*ParseResult)
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

		trees[file.Language] = append(trees[file.Language], file)

		return nil
	})
	if err != nil {
		return raisedIssues, err
	}

	reportFunc := func(pass *Pass, node *sitter.Node, message string) {
		raisedIssues = append(raisedIssues, &Issue{
			Id:       &pass.Analyzer.Name,
			Node:     node,
			Message:  message,
			Filepath: pass.FileContext.FilePath,
		})
	}

	for lang, analyzers := range langAnalyzerMap {
		for _, analyzer := range analyzers {
			allFiles := trees[lang]
			if len(allFiles) == 0 {
				continue
			}

			for _, file := range allFiles {
				pass := &Pass{
					Analyzer:    analyzer,
					FileContext: file,
					Files:       trees[lang],
					Report:      reportFunc,
				}

				_, err := analyzer.Run(pass)
				if err != nil {
					return raisedIssues, err
				}
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
