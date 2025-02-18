package testrunner

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"
)

type Issue struct {
	Column  int
	Message string

	// Hit is used internally by the diff tool to check for issues that were not hit.
	Hit bool
}

// Pragma is of the form [IssueCode]: Column "Message", Column "Message"; [AnotherIssueCode]: ...
type Pragma struct {
	Issues map[string][]*Issue
	Hit    map[string]bool
}

// merge merges another pragma into the same pragma.
func (p *Pragma) merge(other *Pragma) {
	for issueCode, issues := range other.Issues {
		if _, ok := p.Issues[issueCode]; ok {
			p.Issues[issueCode] = append(p.Issues[issueCode], issues...)
		} else {
			p.Issues[issueCode] = issues
		}

		p.Hit[issueCode] = p.Hit[issueCode] || other.Hit[issueCode]
	}
}

var (
	pragmaRegex = regexp.MustCompile(`\s*\[([\w-]+)](:\s*(.*))?`)
	issueRegex  = regexp2.MustCompile(`\s*(\d+)?\s*("(.*?(?<!\\))")?`, regexp2.None)
)

// ParsePragma expects a pragma comment (without the comment prefix like // or #) and
// returns a structured Pragma.
func ParsePragma(comment string) *Pragma {
	if !isPragma(comment) {
		return nil
	}

	pragmaIssues := splitWithEscaping(comment, ";", "\\")
	result := &Pragma{
		Issues: make(map[string][]*Issue),
		Hit:    make(map[string]bool),
	}

	for _, comment := range pragmaIssues {
		pragmaMatches := pragmaRegex.FindAllStringSubmatch(comment, -1)
		if pragmaMatches == nil || len(pragmaMatches) < 1 {
			continue
		}

		pragma := pragmaMatches[0]
		if len(pragma) < 4 {
			continue
		}

		issueCode := pragma[1]

		// Empty column/message pairs
		if pragma[3] == "" {
			result.Issues[issueCode] = []*Issue{}
			result.Hit[issueCode] = false
			continue
		}

		issueCSV := pragma[3]
		for {
			match, err := issueRegex.FindStringMatch(issueCSV)
			if err != nil {
				break
			}

			if match.Group.String() == "" {
				break
			}

			groups := match.Groups()
			if len(groups) != 4 {
				break
			}

			message := groups[2].String()
			if message != "" {
				message, err = strconv.Unquote(message)
				if err != nil {
					break
				}
			}

			issue := &Issue{Message: message, Column: 0}
			if column, err := strconv.Atoi(groups[1].String()); err == nil {
				issue.Column = column
			}

			result.Issues[issueCode] = append(result.Issues[issueCode], issue)
			result.Hit[issueCode] = false

			issueCSV = issueCSV[match.Length:]
			issueCSV = strings.TrimSpace(issueCSV)
			issueCSV = strings.TrimPrefix(issueCSV, ",")
		}
	}

	if len(result.Issues) == 0 {
		return nil
	}

	return result
}

func isPragma(comment string) bool { return pragmaRegex.MatchString(comment) }

func splitWithEscaping(s, separator, escapeString string) []string {
	a := strings.Split(s, separator)

	for i := len(a) - 2; i >= 0; i-- {
		if strings.HasSuffix(a[i], escapeString) {
			a[i] = a[i][:len(a[i])-len(escapeString)] + separator + a[i+1]
			a = append(a[:i+1], a[i+2:]...)
		}
	}
	return a
}
