# Quickstart

Get Globstar on your system, and run your first custom checker in <5 minutes.

## Installation

To install Globstar with [homebrew](https://brew.sh/):

```bash
brew install globstar
```

Or, install with [curl](https://curl.se/) on any Unix-like system:

```bash
curl -sSL https://get.globstar.dev | sh
```

This script will download the latest release of Globstar and install it to `/usr/local/bin`. You can also specify a different installation directory by setting the `BINDIR` environment variable:

```bash
BINDIR=/.local/bin curl -sSL https://get.globstar.dev | sh
```

## Running pre-defined checkers

Globstar comes with a set of pre-defined checkers that you can run on your codebase. Go to the root of a repository and run:

```bash
globstar check --pre-defined
```

This will run all the checkers across all the files in your codebase and print the results to the console. 

## Write your first checker

Create a new folder for Globstar in your repository's root (use a JavaScript project for this example, if you have one — we'll be writing a checker for JavaScript):

```bash
mkdir .globstar
```

Globstar will discover all the checkers in this folder and run them on your codebase. Create a new file in this folder:

```bash
touch .globstar/no_debugger.yml
```

This will define a new checker with the identifier `no_debugger`. Add the following content to the file:

```yaml
# .globstar/no_debugger.yml

language: js 
name: js_no_debugger 
message: Remove the debugger statement before committing your code

# Capture name must be the same as the checker's `name` field.
# Nested captures are allowed to have arbitrary names
pattern: (debugger_statement) @js_no_debugger 

filters:
  # This will not flag any debugger statements
  # nested inside catch blocks
  - pattern-not-inside: (catch_clause)
  # Only match debugger statements inside a function
  # (or some other node that is inside a function)
  - pattern-inside:     (function_declaration)

# files matching these patterns are excluded from analysis
exclude:
  - test/*
  - build/*
  - bin/*

description: |
  The `debugger` statement is a debugging tool that is not meant to be committed to the repository. It can cause the application to stop unexpectedly and is generally considered bad practice. Remove the `debugger` statement before committing your code.
```

This checker will flag any `debugger` statements in your codebase. Run the checker with:

```bash
globstar check
```

That's it! You just ran your first custom checker with Globstar. All in under 5 minutes.