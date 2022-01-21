use std::any::Any;

use crate::RUBY;

use aspen::{
    build_query,
    tree_sitter::{Node, Query, QueryCursor},
    Diagnostic, Lint, LintBuilder,
};
use lazy_static::lazy_static;

lazy_static! {
    pub static ref QUERY: Query = build_query(
        *RUBY,
        r#"
        (
            (assignment left: (_) @lhs right: (_) @rhs) @raise
            (eq? @lhs @rhs)
        )
        "#
    );
    pub static ref LINT: Lint = LintBuilder::default()
        .name("sussign")
        .code("RB-W1002")
        .query(&*QUERY)
        .validate(validator)
        .build()
        .unwrap();
}

fn validator<'a>(
    meta: &Lint,
    node: Node<'a>,
    _ctx: &Option<Box<dyn Any>>,
    src: &[u8],
) -> Vec<Diagnostic> {
    let mut query_cursor = QueryCursor::new();

    let raise_capture = meta.query.capture_index_for_name("raise").unwrap();

    query_cursor
        .matches(&meta.query, node, src)
        .map(|m| {
            let raise = m
                .captures
                .iter()
                .find(|c| c.index == raise_capture)
                .unwrap();
            let at = raise.node.range();
            let text = raise.node.utf8_text(src).unwrap();
            let message = format!("This assignment: `{}`, is suspicious", text);
            Diagnostic::new(at, message)
        })
        .collect::<Vec<_>>()
}

#[cfg(test)]
mod tests {
    use super::LINT;

    use crate::RUBY;

    use aspen::Linter;

    fn linter() -> Linter {
        Linter::new(*RUBY).lint(&LINT).comment_str("#")
    }

    #[test]
    fn trivial() {
        linter().test(
            r#"
            a = a
          # ^^^^^ This assignment: `a = a`, is suspicious
            "#,
        )
    }
}
