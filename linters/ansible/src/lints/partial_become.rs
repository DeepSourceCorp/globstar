use crate::{lints::defs::PARTIAL_BECOME, YAML};

use std::ops::Not;

use aspen::{
    tree_sitter::{Node, Query, QueryCursor},
    Context, Occurrence,
};
use once_cell::sync::Lazy;

static QUERY: Lazy<Query> = Lazy::new(|| {
    Query::new(
        *YAML,
        r#"
        (
            (block_mapping_pair key: (_) @key)
            (#match? @key "become_user")
        )
        "#,
    )
    .unwrap()
});

pub fn validate<'a>(node: Node, ctx: &Option<Context<'a>>, src: &[u8]) -> Vec<Occurrence> {
    ctx.as_ref().map_or(Vec::new(), |c| {
        QueryCursor::new()
            .matches(&QUERY, node, src)
            .flat_map(|m| m.captures)
            .flat_map(|capture| {
                c.scope_of(capture.node)
                    .unwrap()
                    .borrow()
                    .local_defs
                    .iter()
                    .any(|def| def.borrow().name == "become")
                    .not()
                    .then(|| {
                        let at = capture.node.range();
                        let message = "Using `become_user` without `become`";
                        PARTIAL_BECOME.raise(at, message)
                    })
            })
            .collect::<Vec<_>>()
    })
}

#[cfg(test)]
mod tests {
    use crate::YAML;

    use aspen::Linter;

    fn linter() -> Linter {
        Linter::new(*YAML)
            .validator(super::validate)
            .scopes(crate::SCOPES)
            .comment_str("#")
    }

    #[test]
    fn trivial() {
        linter().test(
            r#"
            ---
            - name: Run a command as the apache user
              command: somecommand
              become_user: apache
            # ^^^^^^^^^^^ Using `become_user` without `become`
            "#,
        )
    }

    #[test]
    fn no_match() {
        linter().test(
            r#"
            ---
            - name: Run a command as the apache user
              command: somecommand
              become_user: apache
              become: yes
            "#,
        )
    }
}
