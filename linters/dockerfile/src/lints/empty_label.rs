use crate::DOCKERFILE;

use aspen::{
    build_query,
    tree_sitter::{Node, Query, QueryCursor},
    Context, Lint, Occurrence,
};
use lazy_static::lazy_static;

lazy_static! {
    pub static ref QUERY: Query = build_query(
        *DOCKERFILE,
        r#"
        (
            (label_instruction (label_pair key: (_) @key value: (_) @value))
            (#is? @value "")
        )
        "#
    );
}

const EMPTY_LABEL: Lint = Lint {
    name: "empty-label",
    code: "DO-W1000",
};

fn validator<'a>(node: Node, _ctx: &Option<Context<'a>>, src: &[u8]) -> Vec<Occurrence> {
    let key_capture = QUERY.capture_index_for_name("key").unwrap();
    QueryCursor::new()
        .matches(&QUERY, node, |_: Node| std::iter::empty())
        .flat_map(|m| m.captures)
        .filter(|capture| capture.index == key_capture)
        .map(|capture| {
            let at = capture.node.range();
            let text = capture.node.utf8_text(src).unwrap();
            let message = format!("Found empty label: `{}`", text);
            EMPTY_LABEL.raise(at, message)
        })
        .collect::<Vec<_>>()
}

#[cfg(test)]
mod tests {
    use crate::DOCKERFILE;

    use aspen::Linter;

    fn linter() -> Linter {
        Linter::new(*DOCKERFILE)
            .validator(super::validator)
            .comment_str("#")
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
