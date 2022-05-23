mod lints;

use crate::lints::LINTS;

use aspen::{tree_sitter::Language, Linter};
use lazy_static::lazy_static;

lazy_static! {
    pub static ref GO: Language = tree_sitter_go::language();
}

fn main() {
    let linter = Linter::new(*GO)
        .lints(LINTS.to_vec())
        .comment_str("//")
        .scopes(include_str!("scopes.scm"));
    linter.run_analysis();
}
