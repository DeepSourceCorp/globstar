---
outline: 2
---

# Writing Cross-Function Analysis Checkers with Data Flow Graphs

This tutorial guides you through creating a security checker that performs analysis across different functions in a codebase using Globstar's data flow graphs.

## Understanding Cross-Function Analysis

Cross-function analysis tracks how data flows between different functions within your codebase. This is crucial for detecting vulnerabilities where tainted data passes through multiple functions before reaching a vulnerable sink.

## Prerequisites

- Familiarity with Go programming
- Basic understanding of static analysis concepts
- Knowledge of the Globstar checker framework

## Step 1: Define Your Analyzer

Create an analyzer that depends on the `DataFlowAnalyzer`:

```go
var MyInterFunctionChecker = &analysis.Analyzer{
    Name:        "my_interfunction_checker",
    Language:    analysis.LangJs,
    Description: "Detects vulnerable data flows between functions",
    Category:    analysis.CategorySecurity,
    Severity:    analysis.SeverityCritical,
    Requires:    []*analysis.Analyzer{DataFlowAnalyzer},
    Run:         detectVulnerableInterFunctionFlow,
}
```

## Step 2: Access the Data Flow Graph

Your first step is to access the pre-built data flow graph:

```go
func detectVulnerableInterFunctionFlow(pass *analysis.Pass) (interface{}, error) {
    // Get the data flow graph from analysis pass
    dfg := pass.ResultOf[DataFlowAnalyzer].(*DataFlowGraph)
    if dfg == nil {
        return nil, nil
    }
    
    // Define your vulnerability sources and sinks
    vulnerableSinkFunctions := map[string]struct{}{
        "execCommand": struct{}{},
        "evalScript":  struct{}{},
    }
    
    vulnerableSourceFuncs := map[string]struct{}{
        "getUserInput": struct{}{},
        "getQueryParam": struct{}{},
    }
    
    // Continue with analysis...
    return nil, nil
}
```

## Step 3: Identify Tainted Sources

First, locate all functions that could potentially return tainted data:

```go
// Map to track tainted variables and functions
taintedVariables := make(map[*analysis.Variable]struct{})
taintedFunctions := make(map[string]struct{})

// First pass: identify direct sources of tainted data
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
    
    // Check if this is a known source of tainted data
    funcName := getFunctionName(funcNode, pass.FileContext.Source)
    if _, ok := vulnerableSourceFuncs[funcName]; ok {
        if varName := nameNode.Content(pass.FileContext.Source); varName != "" {
            if variable := dfg.ScopeTree.GetScope(node).Lookup(varName); variable != nil {
                taintedVariables[variable] = struct{}{}
            }
        }
    }
})
```

## Step 4: Analyze Function Bodies for Taint Propagation

Once you've identified tainted sources, analyze how taint propagates through function calls:

```go
// Get all function definitions
for funcName, funcDef := range dfg.FuncDefs {
    // Skip already analyzed functions
    if _, alreadyTainted := taintedFunctions[funcName]; alreadyTainted {
        continue
    }
    
    // Check if any function parameter matches a tainted variable
    for _, param := range funcDef.Parameters {
        if _, isTainted := taintedVariables[param]; isTainted {
            // This function receives tainted data as a parameter
            analyzeAndMarkFunctionTaint(funcName, funcDef, dfg, taintedFunctions, taintedVariables)
        }
    }
}
```

## Step 5: Implement the Taint Analysis Logic

Create a recursive function to analyze taint propagation within function bodies:

```go
func analyzeAndMarkFunctionTaint(
    funcName string, 
    funcDef *FunctionDefinition,
    dfg *DataFlowGraph, 
    taintedFunctions map[string]struct{},
    taintedVariables map[*analysis.Variable]struct{}) bool {
    
    // Mark the function as being analyzed to prevent infinite recursion
    taintedFunctions[funcName] = struct{}{}
    
    // Check function body for:
    // 1. Variables assigned from tainted input
    // 2. Return statements that might return tainted data
    // 3. Calls to vulnerable sink functions with tainted data
    
    tainted := false
    if funcDef.Body != nil {
        analysis.Preorder(funcDef.Body, func(node *sitter.Node) {
            // Look for uses of tainted variables in return statements
            if node.Type() == "return_statement" {
                argument := node.ChildByFieldName("argument")
                if argument != nil && argument.Type() == "identifier" {
                    varName := argument.Content(pass.FileContext.Source)
                    if variable := dfg.ScopeTree.GetScope(node).Lookup(varName); variable != nil {
                        if _, isTainted := taintedVariables[variable]; isTainted {
                            // Function returns tainted data
                            tainted = true
                        }
                    }
                }
            }
            
            // Look for assignments that propagate taint
            if node.Type() == "assignment_expression" {
                // Track taint propagation...
            }
            
            // Look for function calls with tainted arguments
            if node.Type() == "call_expression" {
                // Check if calling sinks with tainted data...
            }
        })
    }
    
    return tainted
}
```

## Step 6: Track Return Values

To fully track inter-function taint, we need to follow return values:

```go
// Map variable declarations to function calls that might return tainted data
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
    
    funcName := getFunctionName(funcNode, pass.FileContext.Source)
    if _, isTainted := taintedFunctions[funcName]; isTainted {
        // This variable receives a value from a function that may return tainted data
        if varName := nameNode.Content(pass.FileContext.Source); varName != "" {
            if variable := dfg.ScopeTree.GetScope(node).Lookup(varName); variable != nil {
                taintedVariables[variable] = struct{}{}
            }
        }
    }
})
```

## Step 7: Detect Vulnerabilities in Sink Function Calls

Finally, look for calls to vulnerable sink functions with tainted data:

```go
// Look for calls to vulnerable functions with tainted data
analysis.Preorder(pass, func(node *sitter.Node) {
    if node == nil || node.Type() != "call_expression" {
        return
    }
    
    funcNode := node.ChildByFieldName("function")
    if funcNode == nil {
        return
    }
    
    funcName := getFunctionName(funcNode, pass.FileContext.Source)
    if _, isVulnerable := vulnerableSinkFunctions[funcName]; isVulnerable {
        // Check if any of the arguments is tainted
        args := node.ChildByFieldName("arguments")
        for i := 0; i < int(args.NamedChildCount()); i++ {
            arg := args.NamedChild(i)
            if arg != nil && arg.Type() == "identifier" {
                varName := arg.Content(pass.FileContext.Source)
                if variable := dfg.ScopeTree.GetScope(node).Lookup(varName); variable != nil {
                    if _, isTainted := taintedVariables[variable]; isTainted {
                        // Found a vulnerability!
                        pass.Report(pass, node, "Passing tainted data to vulnerable function")
                        break
                    }
                }
            }
        }
    }
})
```

## Example: Real-World Cross-Function Flow Analysis

Consider this vulnerable code:

```javascript
// User input enters the system
function getUserInput() {
    return document.getElementById('userInput').value;
}

// Data is processed but not sanitized
function processData(input) {
    return "processed: " + input;
}

// Data is used in a vulnerable way
function executeCommand(command) {
    // Dangerous! Using unsanitized input in exec
    exec(command);
}

// The chain of calls that creates the vulnerability
const userInput = getUserInput();
const processedData = processData(userInput);
executeCommand(processedData);
```

Our inter-function analysis would:

1. Identify `getUserInput()` as a source of tainted data
2. Track the tainted data flowing into `processData()` through its parameter
3. Recognize that `processData()` returns tainted data
4. Detect that `executeCommand()` receives this tainted data
5. Flag the call to `exec()` inside `executeCommand()` as vulnerable

## Advanced Techniques

### Handling Callbacks and Higher-Order Functions

JavaScript often uses callbacks and higher-order functions, which require special handling:

```go
// Look for tainted data in callbacks
if node.Type() == "call_expression" {
    // Check if any argument is a function expression or arrow function
    args := node.ChildByFieldName("arguments")
    for i := 0; i < int(args.NamedChildCount()); i++ {
        arg := args.NamedChild(i)
        if arg != nil && (arg.Type() == "function_expression" || arg.Type() == "arrow_function") {
            // Check if the callback uses tainted variables from parent scope
            analyzeFunctionForCapturedTaint(arg, taintedVariables)
        }
    }
}
```

### Handling Method Calls on Objects

Tracking taint through object methods requires special consideration:

```go
// Process member expressions - obj.method()
if funcNode.Type() == "member_expression" {
    objNode := funcNode.ChildByFieldName("object")
    propNode := funcNode.ChildByFieldName("property")
    
    if objNode != nil && objNode.Type() == "identifier" && propNode != nil {
        objName := objNode.Content(pass.FileContext.Source)
        methodName := propNode.Content(pass.FileContext.Source)
        
        // Check if the object is tainted
        if variable := dfg.ScopeTree.GetScope(node).Lookup(objName); variable != nil {
            if _, isTainted := taintedVariables[variable]; isTainted {
                // The object itself is tainted
                // Mark method call as potentially vulnerable
            }
        }
        
        // Check if this is a known unsafe method
        if _, isUnsafeMethod := unsafeMethods[objName+"."+methodName]; isUnsafeMethod {
            // Potential vulnerability
        }
    }
}
```

## Best Practices

1. **Cache Analysis Results**: Use memoization to avoid re-analyzing functions
2. **Handle Recursion**: Implement cycle detection to prevent infinite recursion
3. **Consider All Return Paths**: Analyze all possible return statements in a function
4. **Track Data Transformations**: Consider whether functions sanitize or manipulate data
5. **Set Analysis Boundaries**: Limit analysis depth to maintain performance

## Conclusion

Cross-function analysis is essential for detecting sophisticated vulnerabilities where tainted data flows through multiple functions. By carefully tracking data flow across function calls and returns, you can build powerful checkers that identify complex security issues in JavaScript applications.

Remember that function invocation patterns in JavaScript can be complex due to callbacks, promises, and functional programming techniques. Your analyzer should handle these patterns to minimize false positives and negatives.
