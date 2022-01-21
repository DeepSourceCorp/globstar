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
        [
            (binary
             left: [(true)(false)]
             right: (_))
            (binary
             left: (_)
             right: [(true)(false)])
        ] @useless-cmp
        "#
    );
    pub static ref LINT: Lint = LintBuilder::default()
        .name("useless-cmp")
        .code("RB-W1001")
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
        .matches(&meta.query, node, |_n: Node| std::iter::empty())
        .flat_map(|m| m.captures)
        .map(|capture| {
            let at = capture.node.range();
            let text = capture.node.utf8_text(src).unwrap();
            let message = format!("This comparison: `{}`, is useless", text,);
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
                puts a == true
                   # ^^^^^^^^^ This comparison: `a == true`, is useless
            "#,
        )
    }

    #[test]
    fn all() {
        linter().test(
            r#"
                puts a == false
                   # ^^^^^^^^^^ This comparison: `a == false`, is useless
                puts true == a
                   # ^^^^^^^^^ This comparison: `true == a`, is useless
                puts false == a
                   # ^^^^^^^^^^ This comparison: `false == a`, is useless
            "#,
        )
    }

    #[test]
    fn negation() {
        linter().test(
            r#"
                puts a != false
                   # ^^^^^^^^^^ This comparison: `a != false`, is useless
                puts true != a
                   # ^^^^^^^^^ This comparison: `true != a`, is useless
                puts false != a
                   # ^^^^^^^^^^ This comparison: `false != a`, is useless
            "#,
        )
    }
}
