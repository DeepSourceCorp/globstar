# Supported Languages

Since we're based on [tree-sitter](https://tree-sitter.github.io/tree-sitter/), we can support all the languages that tree-sitter supports. 

Presently, however, we use the wonderful [go-tree-sitter](https://github.com/smacker/go-tree-sitter) by Maxim Sukharev behind the scenes, which gives us first-class support for the following languages:

| Language | Identifier |
|----------|----------------------|
| Bash | `bash`, `sh` |
| C# | `csharp`, `cs` |
| CSS | `css`, `css3` |
| Dockerfile | `docker`, `dockerfile` |
| Elixir | `elixir` |
| Elm | `elm` |
| Go | `go` |
| Groovy | `groovy` |
| HCL | `hcl`, `tf` |
| HTML | `html` |
| Java | `java` |
| JavaScript | `javascript`, `js`, `jsx` |
| Kotlin | `kotlin`, `kt` |
| Lua | `lua` |
| OCaml | `ocaml`, `ml` |
| PHP | `php` |
| Python | `py`, `python` |
| Ruby | `rb`, `ruby` |
| Rust | `rust`, `rs` |
| Scala | `scala` |
| SQL | `sql` |
| Swift | `swift` |
| TOML | `toml` |
| TypeScript | `typescript`, `ts`, `tsx` |
| YAML | `yaml`, `yml` |

In the checker YAML file, you can specify the language you want to run the checker on by setting the `language` field, which should be one of the identifiers listed above.

> [!NOTE]
> You can write checkers for all the languages listed above using either the YAML or the Go interface. Advanced features like imports and scope resolution, cross-file analysis are **only available in the Go interface** and only for a subset of languages at the moment — Python, and JavaScript.