use std::any::Any;

use crate::DOCKERFILE;

use aspen::{
    tree_sitter::{Node, Query, QueryCursor},
    Diagnostic, Lint, LintBuilder,
};
use lazy_static::lazy_static;

lazy_static! {
    pub static ref QUERY: Query = Query::new(
        *DOCKERFILE,
        r#"
        (
            (label_instruction (label_pair key: (_) @key value: (_) @value))
            (#is? @value "")
        )
        "#
    )
    .unwrap();
    pub static ref LINT: Lint = LintBuilder::default()
        .name("empty-label")
        .code("DO-W1000")
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

    let key_capture = meta.query.capture_index_for_name("key").unwrap();

    query_cursor
        .matches(&meta.query, node, |_: Node| std::iter::empty())
        .flat_map(|m| m.captures)
        .filter(|capture| capture.index == key_capture)
        .map(|capture| {
            let at = capture.node.range();
            let text = capture.node.utf8_text(src).unwrap();
            let message = format!("Found empty label: `{}`", text);
            Diagnostic::new(at, message)
        })
        .collect::<Vec<_>>()
}

#[cfg(test)]
mod tests {
    use super::LINT;

    use crate::DOCKERFILE;

    use aspen::Linter;

    fn linter() -> Linter {
        Linter::new(*DOCKERFILE).lint(&LINT).comment_str("#")
    }

    #[test]
    fn trivial() {
        linter().test(
            r#"
            FROM abc
            LABEL aha=""
                # ^^^ Found empty label: `aha`
            LABEL abc=""
                # ^^^ Found empty label: `abc`
            "#,
        )
    }
}
