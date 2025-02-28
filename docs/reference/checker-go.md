# Checker Go Interface

When you need more complex analysis or custom logic, you can create a checker using Globstar's Go interface. Just like YAML, you can write these checkers in the `.globstar` directory of your repository, and run them using the `globstar check` command.

## Structure of a checker file

For each checker, create a `<checker_name>.go` file in the `.globstar` directory of your repository, with a `<checker_name>.test.<extension>` file for the test cases. The checker file consists of two main components:

1. An `Analyzer` definition that describes metadata about the checker
2. Analysis function that implement the checking logic, with additional helper functions if needed

## Analyzer definition

The `Analyzer` struct defines the basic metadata for your checker:

```go
var MyChecker = &analysis.Analyzer{
    Name:        string,         // Unique identifier for the checker
    Language:    analysis.Lang*, // Target language (e.g., analysis.LangPy)
    Description: string,         // Human-readable description of the issue
    Category:    analysis.Category*, // Issue category (e.g., analysis.CategorySecurity)
    Severity:    analysis.Severity*, // Issue severity (e.g., analysis.SeverityCritical)
    Run:         func(pass *analysis.Pass) (interface{}, error), // Analysis function
}
```

### Fields

| Field | Description |
|-------|-------------|
| **Name** | A unique identifier for your checker (e.g., "dangerous_eval") |
| **Language** | The language the checker targets (e.g., `analysis.LangPy`, `analysis.LangJs`, etc.) |
| **Description** | A description of the issue and its potential impact |
| **Category** | The category of the issue (see categories below) |
| **Severity** | The severity level of the issue (see severities below) |
| **Run** | The function that performs the analysis |

### Categories

| Category | Description |
|----------|-------------|
| `analysis.CategorySecurity` | Security vulnerabilities |
| `analysis.CategoryPerformance` | Performance optimizations |
| `analysis.CategoryBugRisk` | Code that could lead to bugs |
| `analysis.CategoryAntipattern` | Patterns that are considered harmful |
| `analysis.CategoryStyle` | Code style issues |

### Severities

| Severity | Description |
|----------|-------------|
| `analysis.SeverityCritical` | Critical issues that must be fixed immediately |
| `analysis.SeverityError` | Error-level issues that should be addressed soon |
| `analysis.SeverityWarning` | Warning-level issues that may impact code quality |
| `analysis.SeverityInfo` | Informational findings with minimal impact |

## Analysis function

The analysis function is where you implement the pattern detection logic:

```go
func checkPattern(pass *analysis.Pass) (interface{}, error) {
    // Walk the AST
    analysis.Preorder(pass, func(node *sitter.Node) {
        // Pattern matching logic here

        // Report issues when found
        pass.Report(pass, node, "Description of the issue")
    })

    return nil, nil
}
```

### Pass object

The `pass` parameter contains the full context for the current analysis:

| Field | Description |
|-------|-------------|
| **Analyzer** | Reference to the checker that's currently running |
| **FileContext** | The parse result for the current file being analyzed |
| **Files** | All parse results for all files in the analysis (for multi-file analysis) |
| **Report** | Function to report issues found during analysis |

The `FileContext` provides information about the current file:

| Field | Description |
|-------|-------------|
| **Ast** | Root node of the Tree-sitter parse tree for the file |
| **Source** | Raw source code of the file (as a byte slice) |
| **FilePath** | Path to the file being analyzed |
| **TsLanguage** | The Tree-sitter language used to parse the file |
| **Language** | The detected language of the file (as a Globstar Language enum) |
| **ScopeTree** | Information about variable scopes (can be nil if scope support isn't available for this language) |

You'll commonly use:

```go
// Access source code
source := pass.FileContext.Source

// Get file path
filePath := pass.FileContext.FilePath

// Report an issue
pass.Report(pass, node, "Description of the issue")

// Check the analyzer's name or other metadata
analyzerName := pass.Analyzer.Name
```

### Node Traversal

The primary method for traversing the AST is the `analysis.Preorder` function:

```go
analysis.Preorder(pass, func(node *sitter.Node) {
    // This function is called for every node in the AST
})
```

### Node Operations

Tree-sitter nodes provide several methods for examining and navigating the AST. Here are the most commonly used operations:

| Method | Description |
|--------|-------------|
| `node.Type()` | Returns the type of the node (e.g., "call_expression", "identifier") |
| `node.NamedChildCount()` | Returns the number of named children (excludes syntax tokens) |
| `node.NamedChild(i)` | Gets the i-th named child node (0-based index) |
| `node.ChildCount()` | Returns the total number of children (including syntax tokens) |
| `node.Child(i)` | Gets the i-th child node, including syntax tokens |
| `node.ChildByFieldName(name)` | Gets a child with a specific field name (e.g., "function", "arguments") |
| `node.Parent()` | Returns the parent node in the AST |
| `node.Content(source)` | Gets the text content of the node from the source |
| `node.Range()` | Gets the range (start and end position) of the node in the source |

For a complete reference of all available methods, see the [go-tree-sitter documentation](https://pkg.go.dev/github.com/smacker/go-tree-sitter), which is used by Globstar for AST operations. Understanding how to traverse and inspect the AST is crucial for writing effective checkers.

### Reporting Issues

When you find an issue, report it using:

```go
pass.Report(pass, node, "Message describing the issue")
```

## Example: Dangerous use of `eval()`

Here's a basic example of a checker that looks for calls to the `eval()` function:

```go
package checkers

import (
    sitter "github.com/smacker/go-tree-sitter"
    "globstar.dev/analysis"
)

var DangerousEval = &analysis.Analyzer{
    Name:        "dangerous_eval",
    Language:    analysis.LangPy,
    Description: "Using eval() with untrusted input can lead to remote code execution.",
    Category:    analysis.CategorySecurity,
    Severity:    analysis.SeverityCritical,
    Run:         checkDangerousEval,
}

func checkDangerousEval(pass *analysis.Pass) (interface{}, error) {
    analysis.Preorder(pass, func(node *sitter.Node) {
        // Check if this is a function call
        if node.Type() != "call" {
            return
        }

        // Get the function being called
        funcNode := node.ChildByFieldName("function")
        if funcNode == nil || funcNode.Type() != "identifier" {
            return
        }

        // Check if the function is eval()
        if funcNode.Content(pass.FileContext.Source) != "eval" {
            return
        }

        // Report the issue
        pass.Report(pass, node, "Dangerous use of eval() detected")
    })

    return nil, nil
}
```

## Advanced Features

### State Tracking

You can maintain state across the analysis by using variables in your analysis function:

```go
func analyzeWithState(pass *analysis.Pass) (interface{}, error) {
    // State variables
    varTracker := make(map[string]bool)

    // First pass to gather data
    analysis.Preorder(pass, func(node *sitter.Node) {
        // Gather data and update varTracker
    })

    // Second pass to use the gathered data
    analysis.Preorder(pass, func(node *sitter.Node) {
        // Use varTracker to inform analysis
    })

    return nil, nil
}
```

### Context Awareness

You can track context during traversal to make your analysis context-sensitive:

```go
func analyzeWithContext(pass *analysis.Pass) (interface{}, error) {
    inSafeContext := false

    analysis.Preorder(pass, func(node *sitter.Node) {
        // Update context based on where we are in the code
        if node.Type() == "function_definition" {
            name := node.ChildByFieldName("name")
            if name != nil && name.Content(pass.FileContext.Source) == "safe_function" {
                inSafeContext = true
                defer func() { inSafeContext = false }()
            }
        }

        // Use context in analysis
        if !inSafeContext && isSomePattern(node) {
            pass.Report(pass, node, "This pattern is unsafe in this context")
        }
    })

    return nil, nil
}
```

### Scope Analysis

For languages that support it, you can use scope information:

```go
func analyzeWithScopes(pass *analysis.Pass) (interface{}, error) {
    // Check if scope information is available
    if pass.FileContext.ScopeTree == nil {
        return nil, nil
    }

    analysis.Preorder(pass, func(node *sitter.Node) {
        // Get the scope of the current node
        scope := pass.FileContext.ScopeTree.GetScope(node)

        if scope != nil {
            // Look up variables in this scope
            variable := scope.Lookup("someVariable")

            if variable != nil {
                // Check variable properties
                if variable.Kind == analysis.VarKindParameter {
                    // Do something with parameter variables
                }
            }
        }
    })

    return nil, nil
}
```

## Testing

To test your checker, create a test file with examples of code that should and should not trigger your checker:

1. Create a file with the pattern: `.globstar/your_checker_name.test.ext` (e.g., `.globstar/dangerous_eval.test.py`)
2. Mark lines that should trigger an error with `# <expect-error>` comments
3. Run `globstar test` to verify your checker works correctly

## Best practices

1. **Start with test cases**: Define test cases before writing your checker
2. **Keep it focused**: Each checker should focus on a specific pattern or issue
3. **Minimize false positives**: Be careful not to flag legitimate code patterns
4. **Provide clear messages**: Error messages should explain both the problem and potential fixes
5. **Handle edge cases**: Consider complex variations of the pattern you're trying to detect
6. **Use helpers**: Extract complex logic into helper functions for better readability
7. **Consider performance**: For large codebases, make your checker as efficient as possible
