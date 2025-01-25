# Checker YAML Interface

Custom checkers in Globstar are defined using YAML files in your `.globstar` directory. Each checker is a separate YAML file that defines what to look for and how to report issues.

## Example Checker

```yaml
# .globstar/no_console_log.yml
language: js
name: js_no_console_log
message: "Avoid using console.log in production code"
category: style
severity: warning

pattern: |
  (
    call_expression
      (member_expression
        object: (identifier) @console (#eq? @console "console")
        property: (property_identifier) @method (#eq? @method "log"))
  ) @js_no_console_log
  
filters:
  - pattern-inside: (function_declaration)
  - pattern-not-inside: (catch_clause)

exclude:
  - "test/**"

description: |
  Using console.log in production code can impact performance and accidentally 
  expose sensitive information. Use a proper logging library instead.
```

## Required Fields

### `language`
- Type: `string`
- Description: The programming language this checker applies to
- Supported values: 
  - JavaScript: `js`, `javascript`
  - TypeScript: `ts`, `typescript`
  - JSX/TSX: `jsx`, `tsx`
  - Python: `py`, `python`
  - Go: `go`
  - Ruby: `rb`, `ruby`
  - Java: `java`
  - And [many more](/supported-languages)

### `name`
- Type: `string`
- Description: Unique identifier for the checker. There should be at least one matching capture in the pattern query with this name.
- Pattern: `[a-z][a-z0-9_]*`
- Example: `js_no_console_log`

### `message`
- Type: `string`
- Description: Message shown when an issue is found
- Supports variable substitution using `@capture_name`
- Example: `"Avoid using console.log, found: @console_log"`

### `pattern`
- Type: `string`
- Description: Tree-sitter query pattern to match
- Must include at least one capture that matches the checker's name
- Alternative: Use `patterns` for multiple patterns

## Optional Fields

### `category`
- Type: `string`
- Default: `"bug-risk"`
- Allowed values:
  - `style`
  - `bug-risk`
  - `antipattern`
  - `performance`
  - `security`

### `severity`
- Type: `string`
- Default: `"error"`
- Allowed values:
  - `critical`
  - `error`
  - `warning`
  - `info`

### `filters`
- Type: `object[]`
- Description: Refine where patterns match using parent node conditions
- Supported filters:
  - `pattern-inside`: Match only if inside this pattern
  - `pattern-not-inside`: Match only if not inside this pattern

### `exclude`
- Type: `string[]`
- Description: Glob patterns for files to exclude
- Example: `["test/**", "**/*.test.js"]`

### `include`
- Type: `string[]`
- Description: Glob patterns for files to include. If specified, only files matching these patterns will be analyzed.
- Example: `["src/**", "lib/**"]`

### `description`
- Type: `string`
- Description: Detailed explanation of the checker
- Supports markdown formatting

## Pattern Writing Guide

Patterns use tree-sitter's query syntax to match AST nodes. Here are the key concepts:

1. Basic pattern structure:
```yaml
name: js_no_console_log
pattern: |
  (
    call_expression
      (member_expression
        object: (identifier) @console
        property: (property_identifier) @method)
  ) @js_no_console_log
```

2. Using predicates to match specific values:
```yaml
# Match only when identifier is "console" and property is "log"
pattern: |
  (
    call_expression
      (member_expression
        object: (identifier) @console (#eq? @console "console")
        property: (property_identifier) @method (#eq? @method "log"))
  ) @js_no_console_log
```

3. Filters for context-aware matching:
```yaml
pattern: |
  (call_expression) @js_no_console_log
filters:
  - pattern-inside: "(function_declaration)"      # Must be inside a function
  - pattern-not-inside: "(catch_clause)"         # But not inside error handling
```

4. Using captures in messages:
```yaml
message: "Found console.@method call"
pattern: |
  (
    member_expression
      object: (identifier) @console (#eq? @console "console")
      property: (property_identifier) @method
  )
```

Common predicates:
- `#eq?`: Exact string match
- `#match?`: Regex pattern match
- `#not-eq?`: String doesn't match
- `#not-match?`: Regex doesn't match

For a deeper dive into tree-sitter queries:
- [Tree-sitter Query Syntax](https://tree-sitter.github.io/tree-sitter/using-parsers/queries/1-syntax.html)
- [Playground](https://tree-sitter.github.io/tree-sitter/7-playground.html)

## Testing Checkers

Every checker can have an associated test file to verify its behavior. The test file should have the same name as the checker but with a `.test` suffix followed by the appropriate file extension.

### Test File Structure

For a checker `no_console_log.yml`, create a corresponding test file:

```js
// .globstar/no_console_log.test.js

function test() {
  // This should be caught inside a function
  // <expect-error>
  console.log("inside function");

  try {
    something();
  } catch (err) {
    // This should NOT be caught (pattern-not-inside: catch_clause)
    console.log(err);
  }
}

```

### Test Directives

- `<expect-error>`: Place this comment above a line that should trigger the checker

### Running Tests

Use the `test` command to run all checker tests:

```bash
globstar test
```

This will:
1. Find all test files in the `.globstar` directory
2. Run the associated checkers against each test file
3. Verify that errors are reported exactly where expected
4. Exit with status code 1 if any test fails

### Test Writing Tips

1. Test both positive and negative cases:
```js
// Should trigger
// <expect-error>
console.log("debug");

// Should not trigger
console.info("info");
```

2. Test edge cases and variations:
```js
// <expect-error>
console.log();  // Empty call

// <expect-error>
console.log(1, "two", {three: 3});  // Multiple arguments

// <expect-error>
console["log"]("computed property");  // Different syntax
```

3. Test filter conditions:
```js
// Should not trigger (not in a function)
console.log("outside");

function test() {
  // <expect-error>
  console.log("inside function");  // Should trigger

  try {
    something();
  } catch (err) {
    console.log(err);  // Should not trigger (in catch clause)
  }
}
```