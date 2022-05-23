use crate::RUBY;

use aspen::{
    tree_sitter::{Node, Query, QueryCursor},
    Context, Diagnostic, Lint, LintBuilder,
};
use lazy_static::lazy_static;

lazy_static! {
    pub static ref QUERY: Query =
        Query::new(*RUBY, "(if (assignment left: (_) right: (_)) @raise)").unwrap();
    pub static ref LINT: Lint = LintBuilder::default()
        .name("assign-instead-of-eq")
        .code("RB-W1000")
        .query(&*QUERY)
        .validate(validator)
        .build()
        .unwrap();
}

fn validator<'a>(
    meta: &Lint,
    node: Node<'a>,
    _ctx: &Option<Context>,
    src: &[u8],
) -> Vec<Diagnostic> {
    let mut query_cursor = QueryCursor::new();

    query_cursor
        .matches(&meta.query, node, |_n: Node| std::iter::empty())
        .flat_map(|m| m.captures)
        .map(|capture| {
            let at = capture.node.range();
            let text = capture.node.utf8_text(src).unwrap();
            let message = format!(
                "Perhaps this assignment `{}` is supposed to be a comparison",
                text,
            );
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
                if a = 2
                 # ^^^^^ Perhaps this assignment `a = 2` is supposed to be a comparison
                  puts "no"
                else
                  puts "hi"
                end
            "#,
        )
    }
}
