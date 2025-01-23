<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./docs/public/img/logo-wordmark-dark.svg">
    <img alt="Globstar" src="./docs/public/img/logo-wordmark.svg" height="64">
  </picture>
  <p align="center">Fast, feature-rich, open-source static analysis toolkit for writing and running code quality and SAST checkers.</p>
  <hr />
</p>

**Globstar** is an open-source static analysis toolkit that enables developers and security engineers to write code analysis rules and run them using a single, portable binary. Built with performance in mind, Globstar leverages [tree-sitter](https://tree-sitter.github.io/tree-sitter/) for powerful AST-based analysis while providing a simple, intuitive interface for writing rules using the tree-sitter query language. More sophisticated lints can be written using the Go API, which provides access to the tree-sitter AST and the ability to resolve imports and scopes.

Write all your checker rules in the `.globstar` directory for your repository and use the `globstar check` command to run them all across your codebase. It's as simple as that!

## Key Features

* **Lighning-fast**: Written in Go, Globstar is designed to be fast and efficient, making it suitable for large codebases. It's distributed as a single binary, so you don't need to worry about dependencies.

* **Tree-sitter integration**: Write rules using tree-sitter's [S-expressions](https://tree-sitter.github.io/tree-sitter/using-parsers/queries/1-syntax.html) instead of learning a custom DSL. For more sophisticated rules, you can write them in Go using tree-sitter's Go bindings — with multi-file support, import and scope resolution, and more.

* **CI-friendly**: Run Globstar in any CI/CD pipeline by downloading the binary. There are no dependencies to install. It'll automatically detect the `.globstar` directory and run all the rules.

* **Truly open-source**: The Globstar CLI and all pre-defined rules are distributed under the MIT license, so you can use it in your commercial projects without any restrictions. We're committed to ensuring that the project remains available long-term and actively involves the community in its development.

## Getting Started

### Installation

```bash
# using homebrew
brew install globstar

# direct binary download
curl -sSL https://get.globstar.dev | sh
```

Once installed, you can run `globstar check` in your repository to run all the rules in the repository's `.globstar` directory.

### Writing a Rule

Here's an example rule that disallows the `debugger` statement in JavaScript files:

```yml
# .globstar/no_debugger.yml
language: js 
name: js_no_debugger 
message: Remove the "debugger statement" before committing your code

# Capture name must be the same as the rule's `name` field.
# Nested captures are allowed to have arbitrary names

pattern: (debugger_statement) @js_no_debugger 

filters:
  # This will not flag any debugger statements nested inside catch blocks
  - pattern-not-inside: (catch_clause)
  # Only match debugger statements inside a function (or some other node that is inside a function)
  - pattern-inside:     (function_declaration)

exclude: # files matching these patterns are excluded from analysis
  - test/*
  - build/*
  - bin/*

description: |
  The `debugger` statement is a debugging tool that is not meant to be committed to the repository.
  It can cause the application to stop unexpectedly and is generally considered bad practice.
  Remove the `debugger` statement before committing your code.
```

A guide to writing tree-sitter queries can be found [here](https://tree-sitter.github.io/tree-sitter/using-parsers/queries/index.html), along with [this interactive playground](https://tree-sitter.github.io/tree-sitter/7-playground.html).

### Running in CI

To run Globstar in your CI/CD pipeline, simply download the binary and run the `globstar check` command.

Here's an example using GitHub Actions:

```yaml
name: Run Globstar Analysis

on:
  pull_request:
  push:
    branches: [ main ]

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
            
      - name: Install Globstar
        run: curl -sSL https://get.globstar.dev | sh

      - name: Run Globstar checks
        run: globstar check
```

## Why Globstar?

Every team has unique needs, but enforcing custom quality and security has always been hard — either you compromise with generic rules or invest heavily in building custom tooling. At DeepSource, we've spent 5+ years building and perfecting static analyzers that process millions of lines of code daily. This experience taught us exactly what teams need in a custom rules engine — it has to be fast, easy to write, and not require a PhD in compilers.

That's why we built Globstar. Instead of learning a new domain-specific language, you can write rules using tree-sitter's familiar query syntax. If you need more power, write rules in Go with full access to the AST.

We've also made Globstar truly open source under the MIT license. Your tools shouldn't come with strings attached. You can use Globstar freely, modify it to your needs, and trust that it'll remain available. No license gotchas, no forced upgrades.

And if you're using [DeepSource](https://deepsource.com), Globstar will integrate seamlessly in the future — letting you run your custom rules alongside our proprietary analyzers. But that's optional. Globstar stands on its own as a powerful, open-source tool for teams that need to enforce custom standards.


## Documentation

You can find the full documentation for Globstar on the [official website](https://globstar.dev).
