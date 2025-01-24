# Introduction

Welcome to Globstar! Learn everything you need to know on how to use the toolkit to write and run checkers, and use the command-line interface.

## What is Globstar?

Globstar is a static analysis toolkit that helps you write custom checkers to analyze your codebase and detect issues (code quality, security, or whatever you want to check). 

You write checkers in the `.globstar` folder of your repository as in YAML format, with tree-sitter's S-expression syntax to match patterns, and run `globstar check` to run the checkers against your codebase.

For advanced checkers, you can use the Go API and write arbitrarily complex Go code, with full access to the tree-sitter AST and features like imports resolution, scopes, cross-file references, and more.

Whether you are a developer, a security researcher, or a code reviewer, Globstar is a powerful tool to help you automate code analysis and enforce best practices in your codebase.

## Key features

* **Lightning-fast**: Written in Go, Globstar is designed to be fast and efficient, making it suitable for large codebases. It's distributed as a single binary, so you don't need to worry about dependencies.

* **Tree-sitter integration**: Write checkers using tree-sitter's [S-expressions](https://tree-sitter.github.io/tree-sitter/using-parsers/queries/1-syntax.html) instead of learning a custom DSL. For more sophisticated checkers, you can write them in Go using tree-sitter's Go bindings — with multi-file support, import and scope resolution, and more.

* **CI-friendly**: Run Globstar in any CI/CD pipeline by downloading the binary. There are no dependencies to install. It'll automatically detect the `.globstar` directory and run all the checkers.

* **Truly open-source**: The Globstar CLI and all pre-defined checkers are distributed under the MIT license, so you can use it in your commercial projects without any restrictions.

## Why tree-sitter?

Static analysis tools often create their own domain-specific languages for writing rules. While these DSLs can seem simpler at first, they inevitably hit limitations when you need to write complex, precise rules. Worse, you're dependent on the DSL maintainers to keep up with tree-sitter's evolving capabilities and language support. 

That's why Globstar uses tree-sitter's native query syntax — yes, it has a learning curve, but you get direct access to the same powerful pattern-matching language used by GitHub, Neovim, and other major tools. No abstraction layers, no limitations, just pure AST matching power.

Tree-sitter queries map directly to your code's AST structure, which means your rules work exactly as you'd expect them to. When you need to debug a rule, you're working with the actual structure of your code, not an abstraction that might hide important details. 

And the best part: since tree-sitter is widely adopted, modern AI coding assistants are already excellent at generating and modifying these queries. This means you get industrial-strength pattern matching with AI-assisted rule writing that actually understands your code's structure. The learning curve becomes a non-issue, in our humble opinion.


## Where to go next?

- [Installation](/quickstart): Install Globstar on your machine and try running the pre-defined checkers.

- [Writing a Checker](/writing-a-checker): Learn how to write a simple checker in the `.globstar` directory of your repository.

- [Integrating with CI/CD](/ci-cd-integration): Learn how to integrate Globstar with your CI/CD pipeline to run checkers on every commit.

- [Contributing](/contributing): Contribute to Globstar by writing new checkers, improving the CLI, or reporting bugs.