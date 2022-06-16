mod lints;

use crate::lints::LINTS;

use aspen::{tree_sitter::Language, Linter};
use once_cell::sync::Lazy;

pub static RUBY: Lazy<Language> = Lazy::new(|| tree_sitter_ruby::language());

fn main() {
    let linter = Linter::new(*RUBY)
        .validators(LINTS.to_vec())
        .scopes(include_str!("scopes.scm"))
        .comment_str("#");

    if let Err(e) = linter.run_analysis() {
        eprintln!("{}", e)
    }
}
