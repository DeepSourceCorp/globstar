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
        [
            (entrypoint_instruction (shell_command))
            (cmd_instruction (shell_command))
        ] @raise
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

    query_cursor
        .matches(&meta.query, node, src)
        .flat_map(|m| m.captures)
        .map(|capture| {
            let at = capture.node.range();
            let message = "Prefer JSON array format instead of shell command";
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
            ENTRYPOINT some_cmd
          # ^^^^^^^^^^^^^^^^^^^ Prefer JSON array format instead of shell command

            ENTRYPOINT ["some_cmd"]

            CMD some_cmd
          # ^^^^^^^^^^^^ Prefer JSON array format instead of shell command

            CMD ["some_cmd"]
            "#,
        )
    }
}
