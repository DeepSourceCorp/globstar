package analysis

import (
	"fmt"
	"os"
	"path/filepath"

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
	for _, file := range pass.Files {
		walkTree(file.Ast, fn)
	}
}

func RunAnalyzers(path string, analyzers []*Analyzer, format string) error {
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
			return nil
		}

		fmt.Println("Analyzing", path)
		file, err := ParseFile(path)
		if err != nil {
			fmt.Println("Error parsing file", path, err)
			return nil
		}

		trees[file.Language] = append(trees[file.Language], file)

		return nil
	})
	if err != nil {
		return err
	}

	raisedIssues := []*Issue{}
	reportFunc := func(pass *Pass, node *sitter.Node, message string) {
		raisedIssues = append(raisedIssues, &Issue{
			Id:      &pass.Analyzer.Name,
			Node:    node,
			Message: message,
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
					return err
				}
			}
		}
	}

	output, err := reportIssues(raisedIssues, format)
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(output)
	return err
}

func reportIssues(issues []*Issue, format string) ([]byte, error) {
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
