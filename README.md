## OneLint

> One linter, all languages.

A lightweight, extensible linter for all programming languages powered by tree-sitter.

## All languages are equal

But some more than others.
For syntax-based rules that only require an AST, all languages are equally supported in the API.

For a deeper understanding of source code, OneLint has a `LanguageServices` features that extends the capabilities of the linter for specific languages.
This can be used for import resolution, preparing call graphs, and scope analysis.

As an example:

```go
filePath := pyServices.FilePathOfImport(import_node)
```

