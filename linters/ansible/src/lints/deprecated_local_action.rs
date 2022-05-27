use crate::YAML;

use aspen::{
    tree_sitter::{Node, Query, QueryCursor},
    Context, Lint, Occurrence,
};
use lazy_static::lazy_static;

const DEPRECATED_LOCAL_ACTION: Lint = Lint {
    name: "deprecated local action",
    code: "YML-W1002",
};

lazy_static! {
    static ref QUERY: Query = Query::new(
        *YAML,
        r#"
        (
            (block_mapping_pair key: (_) @key)
            (#match? @key "local_action")
        )
        "#
    )
    .unwrap();
}

pub fn validate<'a>(node: Node, _ctx: &Option<Context<'a>>, src: &[u8]) -> Vec<Occurrence> {
    QueryCursor::new()
        .matches(&QUERY, node, src)
        .flat_map(|m| m.captures)
        .map(|capture| {
            let at = capture.node.range();
            let message = "Prefer `delegate_to: localhost` instead of `local_action`";
            DEPRECATED_LOCAL_ACTION.raise(at, message)
        })
        .collect::<Vec<_>>()
}

#[cfg(test)]
mod tests {
    use crate::YAML;

    use aspen::Linter;

    fn linter() -> Linter {
        Linter::new(*YAML)
            .validator(super::validate)
            .comment_str("#")
    }

    #[test]
    fn trivial() {
        linter().test(
            r#"
            ---
            tasks:
              - name: Take out of load balancer pool
                local_action: ansible.builtin.command /usr/bin/take_out_of_pool {{ inventory_hostname }}
              # ^^^^^^^^^^^^ Prefer `delegate_to: localhost` instead of `local_action`   

              - name: Add back to load balancer pool
                local_action: ansible.builtin.command /usr/bin/add_back_to_pool {{ inventory_hostname }}
              # ^^^^^^^^^^^^ Prefer `delegate_to: localhost` instead of `local_action`   
            "#,
        )
    }
}
