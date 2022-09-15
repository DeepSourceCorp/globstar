The `examples` directory contains the following example
analyzers, in increasing order of complexity:

- `simple`: a basic `globstar` analyzer targeting Rust
- `scopes`: an example that covers scope-resolution and
  scoping queries
- `injections`: an example that covers language injections
  using multiple tree-sitter grammars

To run the tests for any example:

```shell
# cargo test --features testing --example <example-name>
$ cargo test --features testing --example simple

# see /globstar/Cargo.toml for a full list of examples
```
