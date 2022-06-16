use crate::{lints::defs::SUSSIGN, RUBY};

use aspen::{
    tree_sitter::{Node, Query, QueryCursor},
    Context, MapCapture, Occurrence,
};
use once_cell::sync::Lazy;

static QUERY: Lazy<Query> = Lazy::new(|| {
    Query::new(
        *RUBY,
        r#"
        (
            (assignment left: (_) @lhs right: (_) @rhs) @raise
            (eq? @lhs @rhs)
        )
        "#,
    )
    .unwrap()
});

pub fn validate<'a>(node: Node, _ctx: &Option<Context<'a>>, src: &[u8]) -> Vec<Occurrence> {
    QueryCursor::new()
        .matches(&QUERY, node, src)
        .map_capture("raise", |raise| {
            let at = raise.node.range();
            let text = raise.node.utf8_text(src).unwrap();
            let message = format!("This assignment: `{}`, is suspicious", text);
            SUSSIGN.raise(at, message)
        })
}

#[cfg(test)]
mod tests {
    use crate::RUBY;

    use aspen::Linter;

    fn linter() -> Linter {
        Linter::new(*RUBY)
            .validator(super::validate)
            .comment_str("#")
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
