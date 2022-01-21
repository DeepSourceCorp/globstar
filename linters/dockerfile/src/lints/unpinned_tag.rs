use std::any::Any;

use crate::DOCKERFILE;

use aspen::{
    build_query,
    tree_sitter::{Node, Query, QueryCursor},
    Diagnostic, Lint, LintBuilder,
};
use lazy_static::lazy_static;

lazy_static! {
    pub static ref QUERY: Query = build_query(
        *DOCKERFILE,
        r#"
        (
            (from_instruction (image_spec name: (_) @name tag: (image_tag) @imgtag))
            (#eq? @imgtag ":latest")
        )
        "#
    );
    pub static ref LINT: Lint = LintBuilder::default()
        .name("unpinned-tag")
        .code("DO-W1003")
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

    let name_capture = meta.query.capture_index_for_name("name").unwrap();

    query_cursor
        .matches(&meta.query, node, src)
        .map(|m| {
            let name = m.captures[name_capture as usize];
            let at = name.node.range();
            let text = name.node.utf8_text(src).unwrap();
            let message = format!("Consider pinning the version of `{}` explicitly", text);
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
            FROM abc:latest
               # ^^^ Consider pinning the version of `abc` explicitly
            RUN git clone x.git
            "#,
        )
    }
}
