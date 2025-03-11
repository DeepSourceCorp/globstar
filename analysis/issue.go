package analysis

import (
	"encoding/json"
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
)

type Issue struct {
	// The category of the issue
	Category Category
	// The severity of the issue
	Severity Severity
	// The message to display to the user
	Message string
	// The file path of the file that the issue was found in
	Filepath string
	// (optional) The AST node that caused the issue
	Node *sitter.Node
	// Id is a unique ID for the issue.
	// Issue that have 'Id's can be explained using the `globstar desc` command.
	Id *string
}

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
	Message  string   `json:"message"`
	Range    position `json:"range"`
	Id       string   `json:"id"`
}

func (i *Issue) AsJson() ([]byte, error) {
	issue := issueJson{
		Category: i.Category,
		Severity: i.Severity,
		Message:  i.Message,
		Range: position{
			Filename: i.Filepath,
			Start: location{
				Row:    int(i.Node.Range().StartPoint.Row) + 1, // 0-indexed to 1-indexed
				Column: int(i.Node.Range().StartPoint.Column),
			},
			End: location{
				Row:    int(i.Node.Range().EndPoint.Row) + 1, // 0-indexed to 1-indexed
				Column: int(i.Node.Range().EndPoint.Column),
			},
		},
		Id: *i.Id,
	}

	return json.Marshal(issue)
}

func (i *Issue) AsText() ([]byte, error) {
	return []byte(fmt.Sprintf("%s:%d:%d:%s", i.Filepath, int(i.Node.Range().StartPoint.Row)+1, i.Node.Range().StartPoint.Column, i.Message)), nil
}

func IssueFromJson(jsonData []byte) (*Issue, error) {
	var issue issueJson
	err := json.Unmarshal(jsonData, &issue)
	if err != nil {
		return nil, err
	}

	return &Issue{
		Category: issue.Category,
		Severity: issue.Severity,
		Message:  issue.Message,
		Filepath: issue.Range.Filename,
		Node:     nil,
		Id:       &issue.Id,
	}, nil
}

func IssueAsTextFromJson(jsonData []byte) ([]byte, error) {
	var issue issueJson
	err := json.Unmarshal(jsonData, &issue)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("%s:%d:%d:%s", issue.Range.Filename, issue.Range.Start.Row, issue.Range.Start.Column, issue.Message)), nil
}
