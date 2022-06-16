use crate::{lints::defs::USELESS_CMP, RUBY};

use aspen::{
    tree_sitter::{Node, Query, QueryCursor},
    Context, Occurrence,
};
use once_cell::sync::Lazy;

static QUERY: Lazy<Query> = Lazy::new(|| {
    Query::new(
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
        "#,
    )
    .unwrap()
});

pub fn validate<'a>(node: Node, _ctx: &Option<Context<'a>>, src: &[u8]) -> Vec<Occurrence> {
    QueryCursor::new()
        .matches(&QUERY, node, |_n: Node| std::iter::empty())
        .flat_map(|m| m.captures)
        .map(|capture| {
            let at = capture.node.range();
            let text = capture.node.utf8_text(src).unwrap();
            let message = format!("This comparison: `{}`, is useless", text,);
            USELESS_CMP.raise(at, message)
        })
        .collect::<Vec<_>>()
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
