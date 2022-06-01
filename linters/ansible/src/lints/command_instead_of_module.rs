use std::path::Path;

use crate::{lint_utils::AsText, lints::defs::COMMAND_INSTEAD_OF_MODULE, BASH, YAML};

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
         key: (_) @command-node
         value: (flow_node (plain_scalar) @value-node))
        (#match? @command-node "^(ansible\\.builtin\\.)?(command|shell)")
    )
    "#,
    )
    .unwrap()
});

static BASH_QUERY: Lazy<Query> = Lazy::new(|| {
    Query::new(
        *BASH,
        r#"
        (
            (command name: (_) @command . argument: (_)? @first-arg)
        )
        "#,
    )
    .unwrap()
});

const MODULES: &[(&str, &str)] = &[
    ("apt-get", "apt-get"),
    ("chkconfig", "service"),
    ("curl", "get_url or uri"),
    ("git", "git"),
    ("hg", "hg"),
    ("letsencrypt", "acme_certificate"),
    ("mktemp", "tempfile"),
    ("mount", "mount"),
    ("patch", "patch"),
    ("rpm", "yum or rpm_key"),
    ("rsync", "synchronize"),
    ("sed", "template, replace or lineinfile"),
    ("service", "service"),
    ("supervisorctl", "supervisorctl"),
    ("svn", "subversion"),
    ("systemctl", "systemd"),
    ("tar", "unarchive"),
    ("unzip", "unarchive"),
    ("wget", "get_url or uri"),
    ("yum", "yum"),
];

const EXECUTABLE_OPTIONS: &[(&str, &[&str])] = &[
    ("git", &["branch", "log", "lfs"]),
    ("systemctl", &["set-default", "show-environment", "status"]),
    ("yum", &["clean"]),
    ("rpm", &["--nodeps"]),
];

pub fn validate<'a>(node: Node, ctx: &Option<Context<'a>>, src: &[u8]) -> Vec<Occurrence> {
    let capture_idx = QUERY.capture_index_for_name("value-node").unwrap();

    ctx.as_ref().map_or(Vec::new(), |c| {
        QueryCursor::new()
            .matches(&QUERY, node, src)
            .flat_map(|m| m.captures.iter().find(|c| c.index == capture_idx))
            .flat_map(|capture| {
                c.injected_tree_of(&capture.node).and_then(|injected_tree| {
                    let original_range = injected_tree.original_range;
                    let root_node = injected_tree.tree.root_node();
                    let message = validate_bash(
                        root_node,
                        &src[original_range.start_byte..original_range.end_byte],
                    );
                    message.map(|m| COMMAND_INSTEAD_OF_MODULE.raise(original_range, m))
                })
            })
            .collect()
    })
}

fn validate_bash<'a>(node: Node, src: &[u8]) -> Option<String> {
    let command_idx = BASH_QUERY.capture_index_for_name("command").unwrap();
    let first_arg = BASH_QUERY.capture_index_for_name("first-arg").unwrap();

    QueryCursor::new()
        .matches(&BASH_QUERY, node, src)
        .next()
        .and_then(|m| {
            let command = m.captures.iter().find(|c| c.index == command_idx)?;
            let first_arg = m.captures.iter().find(|c| c.index == first_arg);
            let executable = Path::new(command.as_text(&src)?).file_name()?.to_str()?;

            if let Some(f) = first_arg.and_then(|f| f.as_text(&src)) {
                if EXECUTABLE_OPTIONS
                    .iter()
                    .any(|&(e, opts)| e == executable && opts.iter().any(|&o| o == f))
                {
                    return None;
                }
            }

            MODULES
                .iter()
                .find(|(cmd, _)| cmd == &executable)
                .map(|(_, alternative)| {
                    format!(
                        "Using command `{}` in place of `{}` module",
                        executable, alternative
                    )
                })
        })
}

#[cfg(test)]
mod tests {
    use crate::{BASH, INJECTIONS, SCOPES, YAML};

    use aspen::{Injection, Linter};

    fn linter() -> Linter {
        Linter::new(*YAML)
            .validator(super::validate)
            .scopes(SCOPES)
            .injection(Injection::new(INJECTIONS, *BASH))
            .comment_str("#")
    }

    #[test]
    fn apt_get() {
        linter().test(
            r#"
            - hosts: all
              tasks:
                - name: run apt-get update
                  command: apt-get update
                         # ^^^^^^^^^^^^^^ Using command `apt-get` in place of `apt-get` module
            "#,
        );
    }

    #[test]
    fn git_no_match() {
        linter().test(
            r#"
            - hosts: all
              tasks:
                - name: print current git branch
                  command: git branch
                - name: print git log
                  command: git log
                - name: install git lfs support
                  command: git lfs install
            "#,
        );
    }

    #[test]
    fn systemd() {
        linter().test(
            r#"
            - hosts: all
              tasks:
                - name: restart sshd
                  command: systemctl restart sshd
                         # ^^^^^^^^^^^^^^^^^^^^^^ Using command `systemctl` in place of `systemd` module
            "#,
        );
    }

    #[test]
    fn systemd_no_match() {
        linter().test(
            r#"
            - hosts: all
              tasks:
                - name: show systemctl service status
                  command: systemctl status systemd-timesyncd
                - name: show systemd environment
                  command: systemctl show-environment
                - name: set systemd runlevel
                  command: systemctl set-default multi-user.target
            "#,
        );
    }

    #[test]
    fn yum() {
        linter().test(
            r#"
            - hosts: all
              tasks:
                - name: run yum update
                  command: yum update
                         # ^^^^^^^^^^ Using command `yum` in place of `yum` module
            "#,
        );
    }

    #[test]
    fn yum_no_match() {
        linter().test(
            r#"
            - hosts: all
              tasks:
                - name: run yum clean
                  command: yum clean
            "#,
        );
    }

    #[test]
    fn shell_script_no_match() {
        linter().test(
            r#"
            - hosts: all
              tasks:
                - name: big script
                  shell: |
                  git checkout master
                  git pull -r
                  git checkout -
                  git rebase master
            "#,
        );
    }
}
