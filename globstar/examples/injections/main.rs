//! A linter that detects usage of `cd` within a `RUN` directive in a Dockerfile,
//! instead of the `WORKDIR` directive.

use globstar::{
    traits::{IsMatch, MapCapture},
    tree_sitter::{Language, Node, Query, QueryCursor},
    Context, Injection, Lint, Linter, Occurrence,
};
use once_cell::sync::Lazy;

static DOCKER: Lazy<Language> = Lazy::new(|| tree_sitter_dockerfile::language());
static BASH: Lazy<Language> = Lazy::new(|| tree_sitter_bash::language());

fn main() {
    // The initialization of our `Linter` is similar to the example provided in `simple/main.rs`,
    // however, we additionally call the `injection` function with an injection query. See
    // the `injections.scm` file for more.
    let linter = Linter::new(*DOCKER)
        .validator(check_run_cd)
        .injection(Injection::new(include_str!("injections.scm"), *DOCKER, *BASH).unwrap());

    if let Err(e) = linter.run_analysis() {
        eprintln!("{e:?}");
    }
}

const RUN_CD: Lint = Lint {
    name: "run-cd",
    code: "DOK-W1000",
};

fn check_run_cd<'a>(node: Node, ctx: &Option<Context<'a>>, src: &[u8]) -> Vec<Occurrence> {
    // this query operates upon the base language (dockerfile)
    let query = Query::new(
        *DOCKER,
        r#"
        (run_instruction (shell_command) @shell)
        "#,
    )
    .unwrap();

    // this query operates upon the target language (bash)
    let bash_query = Query::new(
        *BASH,
        r#"
        (command name: (_) @command-name)
        (#match? @command-name "cd")
        "#,
    )
    .unwrap();

    ctx.as_ref().map_or_else(Vec::new, |ctx| {
        QueryCursor::new()
            .matches(&query, node, src)
            .filter_map_capture("shell", |capture| {
                // Context::injected_tree_of fetches an injected syntaxe tree, if any, at the
                // position of a given node
                ctx.injected_tree_of(&capture.node)
                    .and_then(|injected_tree| {
                        let original_range = injected_tree.original_range;
                        // run any additional processing upon the root node of the injected syntax
                        // tree, in this case, another tree-sitter query
                        let root_node = injected_tree.tree.root_node();
                        QueryCursor::new()
                            .is_match(&bash_query, root_node, src) // see `globstar::traits::IsMatch`
                            .then(|| {
                                let at = original_range;
                                let message = "Use `WORKDIR` directive instead of `RUN cd`";
                                RUN_CD.raise(at, message)
                            })
                    })
            })
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    fn linter() -> Linter {
        Linter::new(*DOCKER)
            .validator(check_run_cd)
            .injection(Injection::new(include_str!("injections.scm"), *DOCKER, *BASH).unwrap())
            .comment_str("#")
    }

    #[test]
    fn smoke_test() {
        linter().test(
            r#"
            RUN cd ..
              # ^^^^^ Use `WORKDIR` directive instead of `RUN cd`
            "#,
        )
    }
}
