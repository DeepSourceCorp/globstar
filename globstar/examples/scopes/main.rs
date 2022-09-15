//! A linter that detects variable shadowing in Elm.

use globstar::{
    traits::MapCapture,
    tree_sitter::{Language, Node, Query, QueryCursor, Range},
    Context, Lint, Linter, Occurrence,
};
use once_cell::sync::Lazy;

static ELM: Lazy<Language> = Lazy::new(|| tree_sitter_elm::language());

fn main() {
    // The initialization of our `Linter` is similar to the example provided in `simple/main.rs`,
    // however, we additionally call the `scopes` function with a scope resolution query. See
    // the `scopes.scm` file for more.
    let linter = Linter::new(*ELM)
        .validator(check_variable_shadowing)
        .scopes(include_str!("scopes.scm"));

    if let Err(e) = linter.run_analysis() {
        eprintln!("{e:?}");
    }
}

const VARIABLE_SHADOWING: Lint = Lint {
    name: "variable-shadowing",
    code: "ELM-W1000",
};

// The idea behind this rule is to look for every definition, and check if any parent
// scope contains a definition by the same name.
fn check_variable_shadowing<'a>(
    node: Node,
    ctx: &Option<Context<'a>>,
    src: &[u8],
) -> Vec<Occurrence> {
    let query = Query::new(
        *ELM,
        r#"
        (function_declaration_left (lower_case_identifier)) @def
        "#,
    )
    .unwrap();

    // Context contains a fully resolved scope tree, created based on the scope query
    ctx.as_ref().map_or_else(Vec::new, |ctx| {
        QueryCursor::new()
            .matches(&query, node, |_n: Node| std::iter::empty())
            .filter_map_capture("def", |capture| {
                let at = capture.node.range();
                let text = capture.node.utf8_text(src).unwrap();

                // `scope_stack_of` produces a stack of scopes, with the current (most deeply
                // nested scope) at the top, and the root (least deeply nested scope) at the bottom
                // of the stack.
                let scope_stack = ctx.scope_stack_of(capture.node).unwrap();
                let mut shadowed_range: Option<Range> = None;

                // check to see if the captured def shadows any pre-existing definition
                // `skip(1)` skips the current scope, and begins the search from the parent
                // scope.
                for scope in scope_stack.skip(1) {
                    for local_def in scope.borrow().local_defs.iter() {
                        if local_def.borrow().name == text {
                            shadowed_range = Some(local_def.borrow().def_range);
                        }
                    }
                }
                if shadowed_range.is_none() {
                    return None;
                }
                let (s_row, s_col) = {
                    let s = shadowed_range.unwrap();
                    (s.start_point.row + 1, s.start_point.column + 1)
                };
                let message = format!(
                    "Shadowing `{}`, defined on line {}, col {}",
                    text, s_row, s_col
                );
                Some(VARIABLE_SHADOWING.raise(at, message))
            })
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    fn linter() -> Linter {
        Linter::new(*ELM)
            .validator(check_variable_shadowing)
            .scopes(include_str!("scopes.scm"))
            .comment_str("--")
    }

    #[test]
    fn smoke_test() {
        linter().test(
            r#"
            f = 2 + 1
            g =
                let
                    a = 2
                    f = 2
                 -- ^ Shadowing `f`, defined on line 1, col 1
                in
                    t + f
            "#,
        )
    }
}
