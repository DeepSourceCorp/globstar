mod lint_utils;
mod lints;

use aspen::{tree_sitter::Language, Injection, Linter};
use lints::LINTS;
use once_cell::sync::Lazy;

pub static YAML: Lazy<Language> = Lazy::new(|| tree_sitter_yaml::language());
pub static BASH: Lazy<Language> = Lazy::new(|| tree_sitter_bash::language());

pub static SCOPES: &str = include_str!("../queries/scopes.scm");
pub static INJECTIONS: &str = include_str!("../queries/injections.scm");

fn main() {
    let linter = Linter::new(*YAML)
        .validators(LINTS.to_vec())
        .comment_str("#")
        .injection(Injection::new(INJECTIONS, *BASH))
        .scopes(SCOPES);

    if let Err(e) = linter.run_analysis() {
        eprintln!("{}", e)
    }
}
