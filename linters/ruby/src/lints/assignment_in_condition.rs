use crate::{lints::defs::ASSIGNMENT_IN_CONDITION, RUBY};

use aspen::{
    tree_sitter::{Node, Query, QueryCursor},
    Context, MapCapture, Occurrence,
};
use once_cell::sync::Lazy;

static QUERY: Lazy<Query> = Lazy::new(|| {
    Query::new(
        *RUBY,
        r#"
        (if
         (assignment
          left: (_) 
          right: (_)) @raise)"#,
    )
    .unwrap()
});

pub fn validate<'a>(node: Node, _ctx: &Option<Context<'a>>, src: &[u8]) -> Vec<Occurrence> {
    QueryCursor::new()
        .matches(&QUERY, node, src)
        .map_capture("raise", |capture| {
            let at = capture.node.range();
            let text = capture.node.utf8_text(src).unwrap();
            let message = format!(
                "Perhaps this assignment `{}` is supposed to be a comparison",
                text,
            );
            ASSIGNMENT_IN_CONDITION.raise(at, message)
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
                if a = 2
                 # ^^^^^ Perhaps this assignment `a = 2` is supposed to be a comparison
                  puts "no"
                else
                  puts "hi"
                end
            "#,
        )
    }
}
