---
outline: 2
---

# Writing a checker in Go

In this guide, we'll use Globstar's Go interface to create a security checker for Python that detects potentially dangerous use of the `eval()` function. We've implemented the same checker [in YAML](/guides/writing-yaml-checker), and this guide will show you how to write it in Go — especially when we're handling some additional complex patterns.

## The dangerous pattern

Let's tackle a serious security issue: use of Python's `eval()` function with untrusted input. For example:

```python
def process_input(user_data):
    result = eval(user_data)  # Dangerous!
    return result
```

This is dangerous because `eval()` can execute arbitrary Python code. An attacker could input malicious code like `"__import__('os').system('rm -rf /')"`. Instead, developers should use safer alternatives like `ast.literal_eval()` for parsing data structures, or proper serialization libraries like `json`.

## Step 1: Writing the test file

Just like with YAML checkers, the best way to start writing a Go checker is to create a comprehensive test file.

Create a file named `.globstar/dangerous_eval.test.py`.

```python
def test_dangerous_eval():
    # These should be flagged
    user_input = get_user_input()

    # <expect-error>
    result1 = eval(user_input)

    # <expect-error>
    result2 = eval("2 + " + user_input)

    # <expect-error>
    result3 = eval(f"2 + {user_input}")

    # These are safe and should not be flagged
    import ast
    safe_result1 = ast.literal_eval('{"name": "test"}')

    # Constants are fine
    safe_result2 = eval("2 + 2")

def test_edge_cases():
    # Should not flag eval in variable names
    evaluation_score = 100

    # Should not flag commented out eval
    # eval(user_input)

    # More complex cases that would be hard to catch with YAML

    def create_evaluator(prefix):
        def evaluate(expr):
            # <expect-error>
            return eval(prefix + expr)  # Nested function with closure
        return evaluate

    class Evaluator:
        def __init__(self, context):
            self.context = context

        def evaluate(self, expr):
            # <expect-error>
            return eval(expr + self.context)  # Method with instance variable
```
Note that we've used the `<expect-error>` annotations right above the line where we expect an error to be raised. This helps us verify that our checker is working correctly.

We've also added some additional complex cases that would be difficult to express in a YAML pattern but can be handled with our Go checker.

## Step 2: Writing the Go checker

Now, let's implement our checker in Go. Create a new file, `.globstar/dangerous_eval.go`:

```go
package checkers

import (
	sitter "github.com/smacker/go-tree-sitter"
	"globstar.dev/analysis"
)

var DangerousEval = &analysis.Analyzer{
	Name:        "dangerous_eval",
	Language:    analysis.LangPy,
	Description: "Using eval() with untrusted input can lead to remote code execution vulnerabilities. Attackers can inject malicious Python code that will be executed by eval().",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityCritical,
	Run:         checkDangerousEval,
}

func checkDangerousEval(pass *analysis.Pass) (interface{}, error) {
	// Walk the AST
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

		// Get the arguments
		argsNode := node.ChildByFieldName("arguments")
		if argsNode == nil {
			return
		}

		// If no arguments, we're safe
		if argsNode.NamedChildCount() == 0 {
			return
		}

		// Get the first argument
		argNode := argsNode.NamedChild(0)
		if argNode == nil {
			return
		}

		// Check if argument is a literal string (usually safe)
		if argNode.Type() == "string" && !containsDynamicContent(argNode, pass.FileContext.Source) {
			return // Safe: eval with constant string
		}

		// Any other pattern is potentially dangerous
		pass.Report(pass, node, "Dangerous use of eval() detected. Use ast.literal_eval() or proper serialization instead.")
	})

	return nil, nil
}

// Helper function to check if a string contains dynamic content like f-strings or concatenation
func containsDynamicContent(node *sitter.Node, source []byte) bool {
	// Check for f-strings (formatted string literals)
	if node.Type() == "string" && len(node.Content(source)) > 0 && node.Content(source)[0] == 'f' {
		return true
	}

	// Check for string concatenation or other dynamic operations
	if node.Type() == "binary_operator" || node.Type() == "comparison_operator" {
		return true
	}

	return false
}
```

Let's break down this code:

### 1. Basic analyzer definition

```go
var DangerousEval = &analysis.Analyzer{
	Name:        "dangerous_eval",
	Language:    analysis.LangPy,
	Description: "Using eval() with untrusted input can lead to remote code execution vulnerabilities...",
	Category:    analysis.CategorySecurity,
	Severity:    analysis.SeverityCritical,
	Run:         checkDangerousEval,
}
```

This defines:
- The unique name for the checker
- The language it applies to
- A description of the issue
- The category and severity of the issue
- The function to run for the analysis

### 2. The analysis function

```go
func checkDangerousEval(pass *analysis.Pass) (interface{}, error) {
	// Walk the AST
	analysis.Preorder(pass, func(node *sitter.Node) {
		// Pattern matching logic here
	})

	return nil, nil
}
```

This function:
1. Takes an analysis pass context that contains the file being analyzed
2. Uses `analysis.Preorder` to visit every node in the AST tree
3. Applies our pattern matching logic
4. Returns any results or errors

### 3. The node matching logic

The core of our checker is the function passed to `analysis.Preorder`. This is where we:

1. Check if the node is a function call
2. Verify the function name is "eval"
3. Examine the arguments to determine if they're safe
4. Report any dangerous usage

### 4. Helper functions

We've added a helper function `containsDynamicContent` to check if a string contains dynamic content like f-strings or string concatenation, which would make it potentially unsafe.


## Step 3: Testing the checker

Run the checker against your test file:

```bash
globstar test
```

If the checker correctly runs and detects the pattern, you should see this in your terminal:

```bash
Running test case: dangerous_eval
All tests passed!
```

## Going beyond YAML checkers

The Go interface for checkers offers several advantages over YAML:

### 1. Complex pattern matching

While YAML checkers are limited to what tree-sitter queries can express, Go checkers can implement arbitrary logic:

```go
// This example shows how we could track variables and follow their flow
func checkDangerousEvalAdvanced(pass *analysis.Pass) (interface{}, error) {
	// Track variables that might contain user input
	unsafeVars := make(map[string]bool)

	// First pass: identify unsafe variables
	analysis.Preorder(pass, func(node *sitter.Node) {
		// Check for assignments from user input functions
		if node.Type() == "assignment" {
			left := node.ChildByFieldName("left")
			right := node.ChildByFieldName("right")

			if isUserInputSource(right, pass.FileContext.Source) && left.Type() == "identifier" {
				unsafeVars[left.Content(pass.FileContext.Source)] = true
			}
		}
	})

	// Second pass: check eval() calls with unsafe variables
	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() == "call" && isFunctionNamed(node, "eval", pass.FileContext.Source) {
			args := node.ChildByFieldName("arguments")
			if args != nil && args.NamedChildCount() > 0 {
				arg := args.NamedChild(0)

				// Check direct variable usage
				if arg.Type() == "identifier" {
					varName := arg.Content(pass.FileContext.Source)
					if unsafeVars[varName] {
						pass.Report(pass, node, "Dangerous use of eval() with potentially tainted input")
					}
				}

				// Check for expressions that use unsafe variables
				if containsAnyVariables(arg, unsafeVars, pass.FileContext.Source) {
					pass.Report(pass, node, "Dangerous use of eval() with expression containing tainted input")
				}
			}
		}
	})

	return nil, nil
}
```

### 2. State tracking

Go checkers can maintain state across the entire analysis:

```go
func checkDangerousEvalWithContext(pass *analysis.Pass) (interface{}, error) {
	// Track the current function context
	inSafeContext := false

	analysis.Preorder(pass, func(node *sitter.Node) {
		// Enter function definition
		if node.Type() == "function_definition" {
			name := node.ChildByFieldName("name")
			if name != nil && name.Content(pass.FileContext.Source) == "safe_eval_wrapper" {
				inSafeContext = true
			}
		}

		// Exit function definition
		if node.Type() == "function_definition" && inSafeContext {
			defer func() { inSafeContext = false }()
		}

		// Check eval calls only outside of safe contexts
		if !inSafeContext && node.Type() == "call" && isFunctionNamed(node, "eval", pass.FileContext.Source) {
			// Report dangerous eval()
			pass.Report(pass, node, "Dangerous use of eval() outside of safe context")
		}
	})

	return nil, nil
}
```

### 3. Type information and semantic analysis

In languages with static typing, Go checkers can leverage type information:

```go
// For a statically-typed language like TypeScript
func checkTypeScriptEvalWithTypeInfo(pass *analysis.Pass) (interface{}, error) {
	// Assuming we have type information available
	typeInfo := pass.TypeInfo

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() == "call_expression" && isFunctionNamed(node, "eval", pass.FileContext.Source) {
			args := node.ChildByFieldName("arguments")
			if args != nil && args.NamedChildCount() > 0 {
				arg := args.NamedChild(0)

				// Get the type of the argument
				argType := typeInfo.TypeOf(arg)

				// Check if the type is safe
				if !isSafeType(argType) {
					pass.Report(pass, node, "Potentially unsafe eval() with non-constant argument")
				}
			}
		}
	})

	return nil, nil
}
```

## Using scope information

Globstar provides scope analysis capabilities for languages that support it:

```go
func checkDangerousEvalWithScopes(pass *analysis.Pass) (interface{}, error) {
	// If scope information is available
	if pass.FileContext.ScopeTree == nil {
		return nil, nil
	}

	analysis.Preorder(pass, func(node *sitter.Node) {
		if node.Type() == "call" && isFunctionNamed(node, "eval", pass.FileContext.Source) {
			args := node.ChildByFieldName("arguments")
			if args != nil && args.NamedChildCount() > 0 {
				arg := args.NamedChild(0)

				// For identifiers, we can check their scope
				if arg.Type() == "identifier" {
					varName := arg.Content(pass.FileContext.Source)

					// Get the scope of this eval() call
					scope := pass.FileContext.ScopeTree.GetScope(node)
					if scope != nil {
						// Look up the variable in this scope
						variable := scope.Lookup(varName)
						if variable != nil {
							// Check if this variable is a parameter or user input
							if variable.Kind == analysis.VarKindParameter {
								pass.Report(pass, node, "Dangerous use of eval() with function parameter")
							}
						}
					}
				}
			}
		}
	})

	return nil, nil
}
```

## Conclusion

Writing checkers in Go gives you tremendous flexibility and power, but requires more code than the YAML approach. It's ideal for complex security checks, semantic analysis, or when you need to track state across an entire file or project. Remember to test your checkers thoroughly with both positive cases (code that should trigger the check) and negative cases (similar code that should not trigger the check) to minimize false positives.
