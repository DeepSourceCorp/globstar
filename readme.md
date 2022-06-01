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

## globstar

light-weight linting with tree-sitter queries. compatible with
`marvin` out-of-the-box.

### architecture

the crates in this workspace are given 5-letter tree names
because nomenclature is hard.

```
[workspace]

members = [
    "marvin",    # rust types/bindings to marvin
    "aspen",     # everything you need to write lints
    "holly",     # annotation based test library
    "cedar",     # scope resolution magic
    "linters/*", # binaries that implement `Linter`s
    "langs/*",   # tree-sitter parsers, mildly vandalized for personal use
]
```

#### `marvin`

the bottommost layer of the stack, contains types, utils and
interfaces to interface with `marvin` gracefully.

#### holly

annotation based test library. extract comments, trim
indents, mark ranges in source code etc:

```ruby
if a = 2
 # ^^^^^ offending code
  puts "hi"
else
  puts "no"
end
```

`holly` contains utils to produce:
- the span of the annotation: `3..7`
- the content of the span: `a = 2`
- the comment: `offending code`

#### `aspen`

provides a "runtime" for lints. `cargo doc --open` for more.
also implements a thin layer over `holly` for pretty
assertions and unit tests. creating `marvin` compatible
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

#### cedar

scope resolution through tree-sitter queries, currently very
primitive and non-configurable.

#### `linters`

this dir contains a couple of sample `aspen::Linter`
implementations:

- `dockerfile`
- `ruby`
