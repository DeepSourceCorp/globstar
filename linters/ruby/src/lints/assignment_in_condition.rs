use std::any::Any;

use crate::RUBY;

use aspen::{
    tree_sitter::{Node, Query, QueryCursor},
    Diagnostic, DiagnosticBuilder, Lint, LintBuilder, Message,
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

fn validator<'a>(meta: &Lint, node: Node<'a>, _ctx: Option<Box<dyn Any>>) -> Vec<Diagnostic> {
    let mut query_cursor = QueryCursor::new();

    query_cursor
        .matches(&meta.query, node, |node: Node| node.kind())
        .flat_map(|m| m.captures)
        .map(|capture| {
            let at = capture.node.range();
            let message = Message::new(
                "Perhaps this assignment `${}` is supposed to be a comparison".into(),
                [at],
            );
            DiagnosticBuilder::default()
                .at(at)
                .message(message)
                .build()
                .unwrap()
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
