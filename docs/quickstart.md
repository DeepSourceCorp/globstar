# Quickstart

Get Globstar on your system, and run your first custom checker in <5 minutes.

## Installation

On a Unix-like system, run:

```bash
curl -sSL https://get.globstar.dev | sh
```

This will download the latest version of Globstar to `./bin/globstar` in your current directory. You can also specify a different installation directory by setting the `BINDIR` environment variable.

```bash
curl -sSL https://get.globstar.dev | BINDIR=$HOME/.local/bin sh
```

Once installed, you can run `./bin/globstar check` or `globstar check` (if installed globally) in your repository to run all the checkers in the builtin checkers that come with Globstar along with all checkers defined in the repository's `.globstar` directory.

## Running builtin checkers

Globstar comes with a set of builtin checkers that you can run on your codebase out-of-the-box. Go to the root of a repository and run:

```bash
globstar check
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

This checker will flag any `debugger` statements in your codebase. 

By default, Globstar will run this checker as well as all the builtin checkers. If you'd like to run only your local checkers:

Run the checker with:

```bash
globstar check --checkers=local
```

That's it! You just ran your first custom checker with Globstar — all in under 5 minutes.