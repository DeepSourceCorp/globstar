mod lints;

use crate::lints::LINTS;

use aspen::{tree_sitter::Language, Linter};
use lazy_static::lazy_static;

lazy_static! {
    pub static ref ELM: Language = tree_sitter_elm::language();
}

fn main() {
    let scopes = include_str!("scopes.scm");
    let linter = Linter::new(*ELM)
        .validators(LINTS.to_vec())
        .comment_str("#")
        .scopes(scopes);
    linter.run_analysis();
}
