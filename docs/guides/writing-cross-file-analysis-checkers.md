---
outline: 2
---

# Writing Cross-File Analysis Checkers with Data Flow Graphs

This tutorial will guide you through creating a security checker that performs analysis across multiple files using Globstar's data flow graphs. We'll use the existing `no_exec` checker as a reference implementation.

## Understanding Data Flow Analysis

Data flow analysis helps track how data moves through your code, which is essential for identifying security vulnerabilities like injection attacks. Cross-file analysis extends this to track data flows across file boundaries.

## Prerequisites

- Familiarity with Go programming
- Basic understanding of static analysis concepts
- Knowledge of the Globstar checker framework

## Step 1: Define Your Analyzer

First, define an analyzer with the correct dependencies:

```go
var MyChecker = &analysis.Analyzer{
    Name:        "my_security_checker",
    Language:    analysis.LangJs,  // Target language
    Description: "Description of the security concern",
    Category:    analysis.CategorySecurity,
    Severity:    analysis.SeverityCritical,
    Requires:    []*analysis.Analyzer{DataFlowAnalyzer},  // Depend on the DataFlowAnalyzer
    Run:         detectVulnerability,
}
```

## Step 2: Design Your Analysis Function

Your main analysis function should:
1. Extract the data flow graph from the analysis pass
2. Define vulnerability sources and sinks
3. Track data flow between them

```go
func detectVulnerability(pass *analysis.Pass) (interface{}, error) {
    // Get the data flow graph from analysis pass
    dfg := pass.ResultOf[DataFlowAnalyzer].(*DataFlowGraph)
    if dfg == nil {
        return nil, nil
    }
    
    // Define your vulnerability sources and sinks
    vulnerableSinkFunctions := map[string]struct{}{
        "dangerousFunction": struct{}{},
    }
    
    vulnerableSourceFuncs := map[string]struct{}{
        "userInput": struct{}{},
    }
    
    // Implement your analysis logic here
    // ...
    
    return nil, nil
}
```

## Step 3: Track Tainted Sources

Identify inputs that could contain malicious content:

```go
// Track potentially tainted inputs
possibleTaintedInputs := make(map[*sitter.Node]string)

// First pass: collect all suspicious source functions
analysis.Preorder(pass, func(node *sitter.Node) {
    if node == nil || node.Type() != "variable_declarator" {
        return
    }
    
    nameNode := node.ChildByFieldName("name")
    valueNode := node.ChildByFieldName("value")
    if nameNode == nil || valueNode == nil || valueNode.Type() != "call_expression" {
        return
    }

    funcNode := valueNode.ChildByFieldName("function")
    if funcNode == nil {
        return
    }

    // Check if this is a source of tainted data
    funcName := getFunctionName(funcNode, pass.FileContext.Source)
    if _, ok := vulnerableSourceFuncs[funcName]; ok {
        possibleTaintedInputs[nameNode] = funcName
    }
})
```

## Step 4: Implement Cross-File Analysis

The key for inter-file analysis is leveraging the `DataFlowGraph` to track variables across files:

```go
func analyze(pass *analysis.Pass) {
    // ...
    // Process the imports in the current file
    if node.Type() == "import_statement" {
        // Handle import statements
        pkg := node.ChildByFieldName("source")
        if pkg == nil && pkg.Type() != "string" {
            return
        }
        pkgName := pkg.Content(pass.FileContext.Source)
        pkgName = filepath.Join(pass.FileContext.FilePath, pkgName)
        if pkgName == "" {
            return
        }
        // check if the import is from a file in the pass.Files.
        // Since the pkgname can be a relative path, with dots,
        // resolve the pkgName to the actual file path
        for _, file := range pass.Files {
            if file.FilePath == pkgName {
                processImportedFile(pass, file)
                break
            }
        }
    }

    // ...
}

func processImportedFile(pass *analysis.File, file *analysis.ParseResult) (any, error) {
    // pick the results from the project-wide cache

	// scope is already built for this file
	scopeTree := pass.ResultCache[ScopeAnalyzer][file].(*analysis.ScopeTree)
	if scopeTree == nil {
		return nil, fmt.Errorf("failed to get scope tree")
	}
	// Check if the DFG is already built
	if pass.ResultCache[pass.Analyzer][file] != nil {
		if dfg, ok := pass.ResultCache[pass.Analyzer][file].(*javascript.DataFlowGraph); ok {
			return dfg, nil
		}
	}

    return nil, fmt.Errorf("failed to get dataflow graph for file")
}
```

## Step 5: Analyze Function Calls for Vulnerabilities

```go
// Track which functions might use tainted data
taintedFunctions := make(map[*FunctionDefinition]struct{})

// Analyze function bodies for vulnerability patterns
for funcName, funcDef := range dfg.FuncDefs {
    if isFunctionVulnerable(funcDef, possibleTaintedInputs) {
        taintedFunctions[funcDef] = struct{}{}
        // Report vulnerability
        pass.Report(pass, funcDef.Node, "Function uses tainted input in an unsafe way")
    }
}
```

## Step 6: Track Function Call Chains

For thorough inter-file analysis, you need to track function call chains:

```go
// Check for tainted data flowing into vulnerable sinks
analysis.Preorder(pass, func(node *sitter.Node) {
    if node == nil || node.Type() != "call_expression" {
        return
    }

    funcNode := node.ChildByFieldName("function")
    if funcNode == nil {
        return
    }
    
    // Check if this is calling a vulnerable sink
    funcName := getFunctionName(funcNode, pass.FileContext.Source)
    if _, ok := vulnerableSinkFunctions[funcName]; ok {
        // Check if any argument is tainted
        args := node.ChildByFieldName("arguments")
        for i := 0; i < int(args.NamedChildCount()); i++ {
            arg := args.NamedChild(i)
            if isArgumentTainted(arg, possibleTaintedInputs, dfg) {
                pass.Report(pass, node, "Potentially unsafe function call with tainted input")
            }
        }
    }
})
```

## Example: Tracing User Input to `exec` Calls

Let's consider an example where we want to detect when user input from one file is passed to an `exec` function in another file:

```javascript
// file1.js
function getUserInput() {
    return window.prompt("Enter command:");
}
module.exports = { getUserInput };

// file2.js
const { getUserInput } = require('./file1');
const input = getUserInput();
exec(input);  // Vulnerability: user input goes directly to exec
```

Our analyzer would:
1. Parse both files
2. Build data flow graphs for each
3. Track the flow of data from `prompt` in file1.js to `exec` in file2.js
4. Report the vulnerability when detected

## Best Practices

1. **Handle Aliasing**: Track when variables are aliased across files
2. **Be Aware of Library Functions**: Know which library functions can be sources or sinks
3. **Avoid False Positives**: Include sanitization function detection
4. **Consider Asynchronous Flows**: Data might flow through Promises or callbacks
5. **Cache Analysis Results**: Reuse analysis results for better performance

## Conclusion

Cross-file analysis is crucial for detecting sophisticated security vulnerabilities that span across multiple files. By leveraging Globstar's data flow capabilities, your checkers can provide deeper insights and uncover issues that would be missed by simpler single-file analysis.

For more details, study the implementation of `no_exec.go` and explore the `DataFlowGraph` structure to understand how to track data across file boundaries effectively.
