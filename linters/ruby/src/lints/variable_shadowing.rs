use crate::{lints::defs::VARIABLE_SHADOWING, RUBY};

use aspen::{
    tree_sitter::{Node, Query, QueryCursor, Range},
    Context, MapCapture, Occurrence,
};
use once_cell::sync::Lazy;

static QUERY: Lazy<Query> = Lazy::new(|| {
    Query::new(
        *RUBY,
        r#"
        (block_parameters (identifier) @def)
        "#,
    )
    .unwrap()
});

pub fn validate<'a>(node: Node, ctx: &Option<Context<'a>>, src: &[u8]) -> Vec<Occurrence> {
    ctx.as_ref().map_or(Vec::new(), |ctx| {
        QueryCursor::new()
            .matches(&QUERY, node, src)
            .flat_map_capture("def", |capture| {
                let at = capture.node.range();
                let text = capture.node.utf8_text(src).unwrap();
                let scope_stack = ctx.scope_stack_of(capture.node).unwrap();
                let mut shadowed_range: Option<Range> = None;
                for scope in scope_stack.skip(1) {
                    for local_def in scope.borrow().local_defs.iter() {
                        if local_def.borrow().name == text {
                            shadowed_range = Some(local_def.borrow().def_range);
                        }
                    }
                }
                let (s_row, s_col) = {
                    let s = shadowed_range?;
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

    use crate::RUBY;

    use aspen::Linter;

    fn linter() -> Linter {
        Linter::new(*RUBY)
            .validator(super::validate)
            .scopes(include_str!("../scopes.scm"))
            .comment_str("#")
    }

    #[test]
    fn trivial() {
        linter().test(
            r#"
            x = 42
            5.times{|x| puts x}
                   # ^ Shadowing `x`, defined on line 1, col 1
            "#,
        )
    }
}
