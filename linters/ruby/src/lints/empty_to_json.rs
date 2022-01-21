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
            (method
             name: (identifier) @n
             !parameters
             )
            (#is? @n "to_json")
        )
        "#
    );
    pub static ref LINT: Lint = LintBuilder::default()
        .name("empty-to-json")
        .code("RB-W1003")
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

    query_cursor
        .matches(&meta.query, node, src)
        .flat_map(|m| m.captures)
        .map(|capture| {
            let at = capture.node.range();
            let message = "This `to_json` method does not accept any parameters";
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
            def to_json
              # ^^^^^^^ This `to_json` method does not accept any parameters
                # some stuff here
            end
            "#,
        )
    }
}
