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
            (workdir_instruction (path) @value)
            (#not-match? @value "^/.*$")
        )
        "#
    )
    .unwrap();
    pub static ref LINT: Lint = LintBuilder::default()
        .name("relative-workdir")
        .code("DO-W1001")
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
            let text = capture.node.utf8_text(src).unwrap();
            let message = format!("Found relative path to WORKDIR directive: `{}`", text,);
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
        dbg!(&LINT.query);
        linter().test(
            r#"
            WORKDIR foo/bar
                  # ^^^^^^^ Found relative path to WORKDIR directive: `foo/bar`
            WORKDIR /abc/def
            "#,
        )
    }
}
