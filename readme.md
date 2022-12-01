# `globstar`

```
      .----------------------------.
      |                            |
      |  ((method                  |
      |    name: (identifier) @fn  |
      |    !parameters)            |
      |   (#is? @fn "to_json"))    |
      |                            |
      `_______________  ,__________'
                      \ |
                       `'  .(()))))).
                          (((((((`)))).
                         ((((`))))))))))
                        )))))))))(((`)(((
                        ))))(((((`)))((((
                       ((((`))))))))()))))
                      (((((((((`)))))))))))
                      ((((((()(((((`)))))))
                       ((((((`)   (((((`))'
                        `""'" |   | "'"'`
                             /     \
                       ~~~~~~~~~~~~~~~~~~~
```

Light-weight linting with tree-sitter queries, compatible with `marvin` out-of-the-box.

## Dependencies

Globstar currently uses a fork of tree-sitter, which is available [here](https://github.com/akshay-deepsource/tree-sitter). You may also require tree-sitter bindings for a specific language during development, e.g., [tree-sitter-go](https://github.com/tree-sitter/tree-sitter-go) or [tree-sitter-ruby](https://github.com/tree-sitter/tree-sitter-ruby).

## Components

The crates in this workspace are given 5-letter tree names because nomenclature is hard.

```
[workspace]

members = [
    "marvin",    # rust types/bindings to marvin
    "aspen",     # everything you need to write lints
    "holly",     # annotation based test library
    "cedar",     # scope resolution magic
    "linters/*", # binaries that implement `Linter`s
]
```
### `marvin`

The bottommost layer of the stack, contains types, utils and
interfaces to interface with `marvin` gracefully.

### holly

Annotation based test library. Extract comments, trim
indents, mark ranges in source code, etc.

```ruby
if a = 2
 # ^^^^^ offending code
  puts "hi"
else
  puts "no"
end
```

`holly` contains utils to produce:
- The span of the annotation: `3..7`
- The content of the span: `a = 2`
- The comment: `offending code`

### `aspen`

Provides a "runtime" for lints. Run `cargo doc --open` for more.
It also implements a thin layer over `holly` for pretty
assertions and unit tests. Creating `marvin` compatible
linters with `aspen` is as simple as:

```rust
let LANGUAGE = tree_sitter_x::language();
let VALIDATORS = vec![ /* a collection of aspen::Validator */ ];

let linter = Linter::new(LANGUAGE).validators(VALIDATORS);

// perform a marvin-compatible analysis run
// - reads analysis_config.json from ANALYSIS_CONFIG_PATH
// - reads file list from CODE_PATH
// - wraps diagnostics into expected serialization format
// - writes analysis_results.json to ANALYSIS_RESULT_PATH
linter.run_analysis()?;
```

### cedar

Scope resolution through tree-sitter queries (currently very
primitive and non-configurable)

### `linters`

This dir contains a couple of sample `aspen::Linter` implementations:

- `dockerfile`
- `ruby`

## Steps to build

You can simply run `cargo build --release`.
