use crate::GO;

use aspen::{
    tree_sitter::{Node, Query, QueryCursor, Range},
    Context, Diagnostic, Lint, LintBuilder,
};
use lazy_static::lazy_static;

lazy_static! {
    pub static ref QUERY: Query = Query::new(
        *GO,
        r#"
        (short_var_declaration
          left: (expression_list
                  (identifier) @def))
        "#
    )
    .unwrap();
    pub static ref LINT: Lint = LintBuilder::default()
        .name("variable-shadowing")
        .code("GO-W1000")
        .query(&*QUERY)
        .validate(validator)
        .build()
        .unwrap();
}

fn validator<'a>(
    meta: &Lint,
    node: Node,
    ctx: &Option<Context<'a>>,
    src: &[u8],
) -> Vec<Diagnostic> {
    let mut query_cursor = QueryCursor::new();

    let ctx = if let Some(c) = ctx { c } else { return vec![] };

    query_cursor
        .matches(&meta.query, node, src)
        .flat_map(|m| m.captures)
        .flat_map(|capture| {
            let at = capture.node.range();
            let text = capture.node.utf8_text(src).unwrap();
            let scope_stack = ctx.scope_stack_of(capture.node).unwrap();
            let mut shadowed_range: Option<Range> = None;
            for scope in scope_stack.skip(1) {
                for local_def in scope.borrow().local_defs.iter() {
                    if local_def.borrow().name == text {
                        shadowed_range = Some(local_def.borrow().def_range);
                    }
                }
            }
            if shadowed_range.is_none() {
                return None;
            }
            let (s_row, s_col) = {
                let s = shadowed_range.unwrap();
                (s.start_point.row + 1, s.start_point.column + 1)
            };
            let message = format!(
                "Shadowing `{}`, defined on line {}, col {}",
                text, s_row, s_col
            );
            Some(Diagnostic::new(at, message))
        })
        .collect::<Vec<_>>()
}

#[cfg(test)]
mod tests {
    use super::LINT;

    use crate::GO;

    use aspen::Linter;

    fn linter() -> Linter {
        Linter::new(*GO)
            .lint(&LINT)
            .scopes(include_str!("../scopes.scm"))
            .comment_str("//")
    }

    #[test]
    fn trivial() {
        linter().test(
            r#"
            func main() {
                i := 10
                {
                    i := 19
                 // ^ Shadowing `i`, defined on line 2, col 5
                }
            }
            "#,
        );
    }
}
