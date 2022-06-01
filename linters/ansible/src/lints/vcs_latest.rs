use crate::{
    lint_utils::{self, AsText},
    lints::defs::{GIT_LATEST, HG_LATEST},
    YAML,
};

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
             key: (_) @vcs-node
             value: 
             (block_node
              (block_mapping
               (block_mapping_pair
                key: (_) @version-node
                value: (_) @version-value) @raise)))
            (#match? @version-node "(version|revision)")
        )
        "#,
    )
    .unwrap()
});

pub fn validate<'a>(node: Node, _ctx: &Option<Context<'a>>, src: &[u8]) -> Vec<Occurrence> {
    let vcs_node_capture = QUERY.capture_index_for_name("vcs-node").unwrap();
    let version_node_capture = QUERY.capture_index_for_name("version-node").unwrap();
    let version_value_capture = QUERY.capture_index_for_name("version-value").unwrap();

    QueryCursor::new()
        .matches(&QUERY, node, src)
        .flat_map(|m| {
            let (mut vcs, mut version_key, mut version_value, mut at) = (None, None, None, None);
            for capture in m.captures {
                let idx = capture.index;
                match idx {
                    _ if idx == vcs_node_capture => vcs = capture.as_text(src),
                    _ if idx == version_value_capture => version_value = capture.as_text(src),
                    _ if idx == version_node_capture => {
                        version_key = capture.as_text(src);
                        at = Some(capture.node.range());
                    }
                    _ => (),
                }
            }
            match (vcs, version_key, version_value, at) {
                (Some(vcs), Some("version"), Some("latest"), Some(at))
                    if lint_utils::is_ansible_builtin(vcs, "git") =>
                {
                    Some(
                        GIT_LATEST
                            .raise(at, "This git checkout does not contain an explicit version"),
                    )
                }
                (Some(vcs), Some("revision"), Some("default"), Some(at))
                    if lint_utils::is_ansible_builtin(vcs, "hg") =>
                {
                    Some(HG_LATEST.raise(
                        at,
                        "This mercurial checkout does not contain an explicit version",
                    ))
                }
                _ => None,
            }
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
            .scopes(crate::SCOPES)
            .comment_str("#")
    }

    #[test]
    fn git() {
        linter().test(
            r#"
            ---
            - git:
                repo: 'https://foosball.example.org/path/to/repo.git'
                dest: /srv/checkout
                version: latest
              # ^^^^^^^ This git checkout does not contain an explicit version
            "#,
        )
    }

    #[test]
    fn mercurial() {
        linter().test(
            r#"
            ---
            - hg:
                repo: 'https://foosball.example.org/path/to/repo.git'
                dest: /srv/checkout
                revision: default
              # ^^^^^^^^ This mercurial checkout does not contain an explicit version
            "#,
        )
    }

    #[test]
    fn no_match() {
        linter().test(
            r#"
            ---
            - git:
                repo: 'https://foosball.example.org/path/to/repo.git'
                dest: /srv/checkout
                version: release-0.22
            "#,
        )
    }
}
