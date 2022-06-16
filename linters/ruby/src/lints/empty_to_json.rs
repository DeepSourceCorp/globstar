use crate::{lints::defs::EMPTY_TO_JSON, RUBY};

use aspen::{
    tree_sitter::{Node, Query, QueryCursor},
    Context, MapCapture, Occurrence,
};
use once_cell::sync::Lazy;

static QUERY: Lazy<Query> = Lazy::new(|| {
    Query::new(
        *RUBY,
        r#"
        (
            (method
             name: (identifier) @n
             !parameters
             )
            (#is? @n "to_json")
        )
        "#,
    )
    .unwrap()
});

pub fn validate<'a>(node: Node, _ctx: &Option<Context<'a>>, src: &[u8]) -> Vec<Occurrence> {
    QueryCursor::new()
        .matches(&QUERY, node, src)
        .map_capture("n", |capture| {
            let at = capture.node.range();
            let message = "This `to_json` method does not accept any parameters";
            EMPTY_TO_JSON.raise(at, message)
        })
}

#[cfg(test)]
mod tests {

    use crate::RUBY;

    use aspen::Linter;

    fn linter() -> Linter {
        Linter::new(*RUBY)
            .validator(super::validate)
            .comment_str("#")
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
