mod lints;

use crate::lints::LINTS;

use aspen::{tree_sitter::Language, Linter};
use lazy_static::lazy_static;

lazy_static! {
    pub static ref DOCKERFILE: Language = tree_sitter_dockerfile::language();
}

fn main() {
    let linter = Linter::new(*DOCKERFILE)
        .lints(LINTS.to_vec())
        .comment_str("#");
    if let Err(e) = linter.run_analysis() {
        eprintln!("error: {}", e);
    }
}
