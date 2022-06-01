use crate::{lints::defs::LITERAL_COMPARE, YAML};

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
            (block_mapping_pair
             key: (_) @conditional-node
             value: [
                ;;; extract plain scalar in `when` param
                ((flow_node) @bin-expr)

                ;;; extract scalar from list of params
                (block_node 
                 (block_sequence 
                  (block_sequence_item 
                   (flow_node) @bin-expr)))
             ])
            (#is? @conditional-node "when")
            (#match? @bin-expr "[=!]= ?(True|true|False|false)")
        )
        "#,
    )
    .unwrap()
});

pub fn validate<'a>(node: Node, _ctx: &Option<Context<'a>>, src: &[u8]) -> Vec<Occurrence> {
    let capture_idx = QUERY.capture_index_for_name("bin-expr").unwrap();
    QueryCursor::new()
        .matches(&QUERY, node, src)
        .flat_map(|m| m.captures.iter().filter(|c| c.index == capture_idx))
        .map(|capture| {
            let at = capture.node.range();
            let message =
                "Consider using `when: var` or `when: not var` instead of boolean comparison";
            LITERAL_COMPARE.raise(at, message)
        })
        .collect()
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
            - name: example task
              debug:
                msg: test
              when: my_var == false
                  # ^^^^^^^^^^^^^^^ Consider using `when: var` or `when: not var` instead of boolean comparison
            "#,
        )
    }

    #[test]
    fn singleton_list() {
        linter().test(
            r#"
            ---
            - name: another example task
              debug:
                msg: test
              when:
                - my_var == false
                # ^^^^^^^^^^^^^^^ Consider using `when: var` or `when: not var` instead of boolean comparison
            "#,
        )
    }

    #[test]
    fn multi_item_list() {
        linter().test(
            r#"
            ---
            - name: another example task
              debug:
                msg: test
              when:
                - my_var == false
                # ^^^^^^^^^^^^^^^ Consider using `when: var` or `when: not var` instead of boolean comparison
                - my_var == True
                # ^^^^^^^^^^^^^^ Consider using `when: var` or `when: not var` instead of boolean comparison
            "#,
        )
    }

    #[test]
    fn no_match_trivial() {
        linter().test(
            r#"
            ---
            - name: another example task
              debug:
                msg: test
              when: my_var
            "#,
        )
    }
    #[test]
    fn no_match_list() {
        linter().test(
            r#"
            ---
            - name: another example task
              debug:
                msg: test
              when:
                - my_var
            "#,
        )
    }
}
