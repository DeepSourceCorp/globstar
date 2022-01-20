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
            (run_instruction (shell_command (shell_fragment) @fragment))
            (#match? @fragment "^cd.*$")
        )
        "#
    )
    .unwrap();
    pub static ref LINT: Lint = LintBuilder::default()
        .name("run-cd")
        .code("DO-W1002")
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
            let message = "Use `WORKDIR` directive instead of `RUN cd`";
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
            RUN cd ..
              # ^^^^^ Use `WORKDIR` directive instead of `RUN cd`
            RUN git clone x.git
            "#,
        )
    }
}
