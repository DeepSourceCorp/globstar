//! A linter that detects the usage of certain function names in
//! Rust code and flags them.

use globstar::{
    traits::MapCapture,
    tree_sitter::{Language, Node, Query, QueryCursor},
    Context, Lint, Linter, Occurrence,
};

use once_cell::sync::Lazy;

static RUST: Lazy<Language> = Lazy::new(|| tree_sitter_rust::language());

fn main() {
    // initalize a linter for the rust language
    let linter = Linter::new(*RUST)
        .extension("rs")
        .ignore(r#"Cargo\.toml"#)
        .validator(check_disallowed_method_names);

    if let Err(e) = linter.run_analysis() {
        eprintln!("{e:?}");
    }
}

// A list of names to be disallowed in Rust code
const DISALLOWED_NAMES: &[&str] = &[
    // avoid placeholder names
    "foo",
    "bar",
    // avoid methods that can cause UB
    "str::from_utf8_unchecked",
];

// Declare a lint to raise
const DISALLOWED_NAMES_RULE: Lint = Lint {
    name: "disallowed-names",
    code: "RS-W1000",
};

// This is our validator function, see `globstar::ValidatorFn` for more. Each "rule" is expressed
// as a `ValidatorFn`. globstar provides a few traits (see `globstar::traits`) to help with the
// construction of a `ValidatorFn`.
//
// `node`: This refers to the `root_node` of every file that is successfully parsed.
// `_ctx`: This contains information about scopes and injections, if available
//   (see other examples for more)
// `src`: Contains the source for the entire file as bytes. This may be passed into
// QueryCursor::matches, if your query contains captures.
fn check_disallowed_method_names<'a>(
    node: Node,
    _ctx: &Option<Context<'a>>,
    src: &[u8],
) -> Vec<Occurrence> {
    // The easiest way to detect an antipattern in source-code is to use a tree-sitter query.
    // This query looks for all function call expressions, and captures the name of the
    // function in the `@function-name` capture.
    let query = Query::new(*RUST, "(call_expression function: (_) @function-name)").unwrap();

    QueryCursor::new()
        .matches(&query, node, src)
        // globstar provides a helper function `filter_map_capture` in the `MapCapture` trait,
        // that applies a closure to every capture by name. In this case, we are raising
        // the DISALLOWED_NAMES_RULE lint upon every capture called `function-name`.
        .filter_map_capture("function-name", |capture| {
            let text = capture.node.utf8_text(src).unwrap();
            let location = capture.node.range();
            DISALLOWED_NAMES.contains(&text).then(|| {
                let message = format!("Use of disallowed method: `{}`", text);
                // Use Lint::raise to produce an `Occurrence`
                DISALLOWED_NAMES_RULE.raise(location, message)
            })
        })
}

// globstar provides a few helpers for fixture-based testing. These are hidden away
// behind the `testing` feature. Use `cargo test --features testing` to make use of
// globstar's test utilites.
#[cfg(test)]
mod tests {
    use super::*;

    // For each test, we setup a linter, not too different from the one used in our `main`
    // function:
    fn setup_linter() -> Linter {
        Linter::new(*RUST)
            .comment_str("//")
            .validator(check_disallowed_method_names)
    }

    #[test]
    fn smoke_test() {
        let linter = setup_linter();

        // To test the lint, call Linter::test, by passing annotated source code as the argument.
        // It is a good practice to put the source code in a raw string. Locations that are
        // expected to be annotated, have to be followed by an "annotation" comment. These comments
        // are single line comments, with the "caret" character to denote the span of the
        // occurrence, and the message provided by the occurrence:
        linter.test(
            r#"
            fn main() {
                let a = foo();
                     // ^^^ Use of disallowed method: `foo`

                unsafe {
                    let b = str::from_utf8_unchecked("bar");
                         // ^^^^^^^^^^^^^^^^^^^^^^^^ Use of disallowed method: `str::from_utf8_unchecked`
                }
            }
            "#,
        )
    }
}
