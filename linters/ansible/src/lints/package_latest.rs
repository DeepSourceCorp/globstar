use crate::{lints::defs::PACKAGE_LATEST, YAML};

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
            (block_mapping_pair
             key: (_) @key
             value: 
             (block_node 
              (block_mapping_pair
               key: (_) @state-key
               value: (_) @version-key)))
            (#match? @state-key "^state")
            (#match? @version-key "^latest")
            (#match? @key "apk|apt|bower|bundler|dnf|easy_install|gem|homebrew|jenkins_plugin|npm|openbsd_package|openbsd_pkg|package|pacman|pear|pip|pkg5|pkgutil|portage|slackpkg|sorcery|swdepot|win_chocolatey|yarn|yum|zypper|")
        )
        "#,
    )
    .unwrap()
});

pub fn validate<'a>(node: Node, ctx: &Option<Context<'a>>, src: &[u8]) -> Vec<Occurrence> {
    let capture_idx = QUERY.capture_index_for_name("key").unwrap();

    ctx.as_ref().map_or(Vec::new(), |c| {
        QueryCursor::new()
            .matches(&QUERY, node, src)
            .flat_map(|m| m.captures.iter().find(|c| c.index == capture_idx))
            .flat_map(|capture| {
                let at = capture.node.range();
                let message = "Consider";
            })
            .collect()
    })
}
