# Configuration Reference

Globstar can be configured using a `.config.yml` file in your repository's `.globstar` directory. You can use this file to set the default behavior for Globstar, including which checkers to run, which directories to analyze, and more.

## Example Configuration

```yaml
# .globstar/.config.yml

ruleDir: .globstar
enabledRules:
  - js_no_debugger
  - py_no_print
disabledRules:
  - js_console_log
targetDirs:
  - src/
  - lib/
excludePatterns:
  - "test/**"
  - "**/*.test.js"
failWhen:
  exitCode: 1
  severityIn:
    - critical
    - error
  categoryIn:
    - security
    - bug-risk
```

::: info FYI 
All configuration fields are optional. Globstar will use sensible defaults if a field is not specified. You can configure just the options you need to customize and leave the rest at their default values.
:::

## Configuration Options

### `ruleDir`
- Type: `string`
- Default: `.globstar`
- Description: Directory containing custom checker definitions

### `enabledRules`
- Type: `string[]`
- Default: All rules
- Description: List of checker IDs to enable. If specified, only these checkers will run.

### `disabledRules`
- Type: `string[]`
- Default: None
- Description: List of checker IDs to disable. These checkers will be skipped during analysis.

### `targetDirs`
- Type: `string[]`
- Default: Current directory
- Description: List of directories to analyze. Useful for monorepos or when you want to analyze specific directories.

### `excludePatterns`
- Type: `string[]`
- Default: Common patterns (see below)
- Description: Glob patterns for files/directories to exclude from analysis.

### `failWhen`
Configuration for when Globstar should exit with a non-zero status code.

#### `exitCode`
- Type: `integer`
- Default: `1`
- Description: Exit code to return when failure conditions are met

#### `severityIn`
- Type: `string[]`
- Default: `["critical"]`
- Allowed values:
  - `critical`
  - `error`
  - `warning`
  - `info`
- Description: List of severities that should trigger a failure

#### `categoryIn`
- Type: `string[]`
- Default: `["bug-risk"]`
- Allowed values:
  - `style`
  - `bug-risk`
  - `antipattern`
  - `performance`
  - `security`
- Description: List of categories that should trigger a failure

## Default Exclusions

By default, Globstar ignores the following directories:
```yaml
- node_modules
- vendor
- dist
- build
- out
- .git
- .svn
- venv
- __pycache__
- .idea
- .globstar
```

## Environment Variables

Configuration can also be influenced by environment variables:

- `GLOBSTAR_CONFIG`: Path to a custom config file
- `GLOBSTAR_DEBUG`: Enable debug logging when set to `true`

## Exit Codes

- `0`: No issues found or issues don't match failure conditions
- `1` (default): Issues found matching failure conditions
- Any non-zero value specified in `failWhen.exitCode`
