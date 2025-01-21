## OneLint

A lightweight tool to write and execute lints with the tree-sitter QL.
More sophisticated lints (e.g: with scope and import resolution) can be written
using the Golang API.

Keep all lints in the  `.one/` directory, and use the `one lint` command
to run them all.

## Adding a new lint

Here's an example lint that disallows the `debugger` statement in JavaScript files:

```yml
# .one/no_debugger.yml
language: js 
name: js_no_debugger 
message: Remove the "debugger statement" before committing your code
# capture name must be the same as the lint's `name` field.
# Nested captures are allowed to have arbitrary names
pattern: (debugger_statement) @js_no_debugger 
exclude: # these files are excluded from analysis
  - test/*
  - build/*
  - bin/*
description: |
  The `debugger` statement is a debugging tool that is not meant to be committed to the repository.
  It can cause the application to stop unexpectedly and is generally considered bad practice.
  Remove the `debugger` statement before committing your code.
```


A guide to writing tree-sitter queries can be found [here](https://tree-sitter.github.io/tree-sitter/using-parsers/queries/index.html), along with [this interactive playground](https://tree-sitter.github.io/tree-sitter/7-playground.html).
