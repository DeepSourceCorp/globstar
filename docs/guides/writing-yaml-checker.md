---
outline: 2
---
# Writing a checker in YAML

In this guide, we'll walk through creating a security checker for Python that detects potentially dangerous use of the `eval()` function. We'll build this step-by-step, starting with test cases and working our way to a complete, working checker. To see all the full specification for writing a checker in YAML, see the [Checker YAML Interface](/reference/checker-yaml) page.

## The dangerous pattern

Let's tackle a serious security issue: use of Python's `eval()` function with untrusted input. For example:

```python
def process_input(user_data):
    result = eval(user_data)  # Dangerous!
    return result
```

This is dangerous because `eval()` can execute arbitrary Python code. An attacker could input malicious code like `"__import__('os').system('rm -rf /')"`. Instead, developers should use safer alternatives like `ast.literal_eval()` for parsing data structures, or proper serialization libraries like `json`.

Let's write a Globstar checker to detect this pattern.

## Step 1: Writing the test file

The best way to start writing a checker is to create a comprehensive test file. This helps you:
1. Think through all the patterns you want to catch
2. Ensure your checker works as expected
3. Avoid false positives

Create a file named `.globstar/dangerous_eval.test.py`.

The format of the filename is `.globstar/<checker_name>.test.py`. The checker name should be meaningful and unique across all checkers in your `.globstar` directory. You will be using this name later in the checker definition.

Now, let's start writing the test cases.
```python
def test_dangerous_eval():
    # These should be flagged
    user_input = get_user_input()

    # <expect-error>
    result1 = eval(user_input)

    # <expect-error>
    result2 = eval("2 + " + user_input)

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
```

Notice how we've:
- Used `<expect-error>` comments to mark lines that should trigger our checker
- Included both positive cases (dangerous patterns) and negative cases (safe patterns)
- Added edge cases to prevent false positives
- Covered different usage patterns

## Step 2: Writing the checker

Now that we have our test cases, let's write the checker. Create `.globstar/dangerous_eval.yml`:

```yaml
language: python
name: dangerous_eval
message: "Dangerous use of eval() detected. Use ast.literal_eval() or proper serialization instead."
category: security
severity: critical

pattern: >
    (call
      function: (identifier) @func
      (#eq? @func "eval")
      arguments: (argument_list
        [
          (identifier)
          (binary_operator)
        ]
      )
    ) @dangerous_eval

filters:
  - pattern-inside: (function_definition)

exclude:
  - "test/**"
  - "**/*_test.py"

description: |
  Using eval() with untrusted input can lead to remote code execution vulnerabilities.
  Attackers can inject malicious Python code that will be executed by eval().
```

Let's break down the key components:

### 1. Basic metadata
```yaml
language: python
name: dangerous_eval
message: "Dangerous use of eval() detected. Use ast.literal_eval() or proper serialization instead."
category: security
severity: critical
```
This defines:
- The language this checker applies to, and should be one of the identifiers defined in [supported languages](/supported-languages)
- A unique identifier for the checker
- The message shown when an issue is found
- The category and severity of the issue

### 2. The pattern
The pattern is where we define what to look for using tree-sitter's query syntax:

```yaml
pattern: >
  (
    call
    function: (identifier) @func
    (#eq? @func "eval")
    arguments: (argument_list
      [
        (identifier)  # Match variable arguments
        (binary_operator)  # Match string concatenation
      ]
    )
  ) @dangerous_eval
```

Let's break this down step by step:

1. `(call ...)`: We're looking for function calls
2. `function: (identifier) @func`: Capture the function name in @func
3. `(#eq? @func "eval")`: Check if that function is "eval"
4. `arguments: (argument_list [... ])`: Look at the arguments list
5. Inside the argument list, we use `[...]` to match either:
   - `(identifier)`: A variable being passed to eval
   - `(binary_operator)`: String concatenation in the argument
6. Finally, `@dangerous_eval` captures the entire dangerous pattern

This pattern will catch:
```python
eval(user_input)              # Matches identifier
eval("2 + " + user_input)     # Matches binary_operator
```

### 3. Filters and exclusions
```yaml
filters:
  - pattern-inside: (function_definition)

exclude:
  - "test/**"
  - "**/*_test.py"
```
- `filters` ensure we only match within functions to reduce false positives
- `exclude` prevents the checker from running on test files

## Step 3: Testing the checker

Run the checker against your test file:

```bash
globstar test
```

If the checker correctly runs and detects the pattern, you should see this in your terminal:

```bash
Running test case: dangerous_eval.yml
All tests passed!
```

## About tree-sitter patterns

The pattern syntax might look intimidating at first, but it's quite logical. Here's how to think about it:

1. **Tree Structure**: Every piece of code has a tree structure. For example:
   ```python
   result = eval(user_input)
   ```
   becomes:
   ```
   assignment
   ├── left: identifier "result"
   └── right: call
       ├── function: identifier "eval"
       └── arguments: argument_list
           └── identifier "user_input"
   ```

2. **Pattern Matching**: Our pattern describes this tree structure:
   ```
   (call
     function: (identifier) @func
     arguments: (argument_list ...))
   ```
   This says "find a function call where the function is an identifier"

3. **Predicates**: The `#eq?` predicate adds additional conditions:
   ```
   (#eq? @func "eval")
   ```
   This ensures we only match calls to the `eval` function

You can use the [tree-sitter playground](https://tree-sitter.github.io/tree-sitter/7-playground.html) to explore how your code is parsed and experiment with patterns.


## Using AI to generate patterns

One of the most powerful ways to write tree-sitter patterns is to leverage Large Language Models (LLMs) like ChatGPT or Claude. Here's an optimized prompt template you can use, with modifications according to your needs:

```
I need help writing a tree-sitter pattern for <PROGRAMMING_LANGUAGE> code. I want to detect this pattern:

<PASTE YOUR TEST FILE CONTENTS HERE>

Please give me:
1. The tree-sitter S-expression pattern that would match this code
2. A brief explanation of how the pattern works
3. A few examples of what it would and wouldn't match

Some requirements:
- The pattern should use named captures with @<YOUR_CHECKER_NAME> syntax
- Include any necessary predicates (#eq?, #match?, etc.)
- Consider common variations of the pattern
- Think about potential false positives
```

For example, for our `eval()` checker, you might use:

```
I need help writing a tree-sitter pattern for Python code. I want to detect dangerous uses of eval() like this:

eval(user_input)
eval("2 + " + user_input)

Please give me:
1. The tree-sitter S-expression pattern that would match this code
2. A brief explanation of how the pattern works
3. A few examples of what it would and wouldn't match

Some requirements:
- The pattern should capture the entire eval call with @py_dangerous_eval
- Include any necessary predicates (#eq?, #match?, etc.)
- Consider common variations of the pattern
- Think about potential false positives
```

### Tips for using AI

1. **Provide Multiple Examples**: Include both positive and negative examples in your prompt
2. **Be Specific**: Mention any particular captures or predicates you need
3. **Verify**: Always test AI-generated patterns in the tree-sitter playground
4. **Iterate**: If the pattern isn't quite right, ask for refinements based on specific issues you find

With these principles in mind, you can write effective checkers for any code pattern you want to detect or enforce in your codebase. Now go, write your own checker!
