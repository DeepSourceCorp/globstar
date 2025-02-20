package analysis

import (
	"os"
	"path/filepath"
	"encoding/json"

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

func RunAnalyzers(path string, analyzers []*Analyzer) ([]*Issue, error) {
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
			return nil
		}

		file, err := ParseFile(path)
		if err != nil {
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
	case "sarif":
		return reportSARIF(issues)
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

func reportSARIF(issues []*Issue) ([]byte, error) {
	type sarifLocation struct {
		Uri string `json:"uri"`
	}

	type sarifRegion struct {
		StartLine   int `json:"startLine"`
		StartColumn int `json:"startColumn"`
		EndLine     int `json:"endLine"`
		EndColumn   int `json:"endColumn"`
	}

	type sarifPhysicalLocation struct {
		ArtifactLocation sarifLocation `json:"artifactLocation"`
		Region           sarifRegion   `json:"region"`
	}

	type sarifResult struct {
		RuleId           string               `json:"ruleId"`
		Message          string               `json:"message"`
		Level            string               `json:"level"`
		PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
	}

	type sarifRun struct {
		Tool struct {
			Driver struct {
				Name           string `json:"name"`
				Version        string `json:"version"`
				InformationUri string `json:"informationUri"`
				Rules          []struct {
					Id          string `json:"id"`
					Name        string `json:"name"`
					ShortDescription struct {
						Text string `json:"text"`
					} `json:"shortDescription"`
					FullDescription struct {
						Text string `json:"text"`
					} `json:"fullDescription"`
					HelpUri string `json:"helpUri"`
				} `json:"rules"`
			} `json:"driver"`
		} `json:"tool"`
		Results []sarifResult `json:"results"`
	}

	type sarifLog struct {
		Version string    `json:"version"`
		Runs    []sarifRun `json:"runs"`
	}

	sarif := sarifLog{
		Version: "2.1.0",
		Runs:    []sarifRun{},
	}

	run := sarifRun{}
	run.Tool.Driver.Name = "Globstar"
	run.Tool.Driver.Version = "1.0.0"
	run.Tool.Driver.InformationUri = "https://globstar.dev"

	for _, issue := range issues {
		result := sarifResult{
			RuleId:  *issue.Id,
			Message: issue.Message,
			Level:   string(issue.Severity),
			PhysicalLocation: sarifPhysicalLocation{
				ArtifactLocation: sarifLocation{
					Uri: issue.Filepath,
				},
				Region: sarifRegion{
					StartLine:   int(issue.Node.Range().StartPoint.Row) + 1,
					StartColumn: int(issue.Node.Range().StartPoint.Column),
					EndLine:     int(issue.Node.Range().EndPoint.Row) + 1,
					EndColumn:   int(issue.Node.Range().EndPoint.Column),
				},
			},
		}

		run.Results = append(run.Results, result)
	}

	sarif.Runs = append(sarif.Runs, run)

	return json.Marshal(sarif)
}
