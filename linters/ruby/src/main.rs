mod lints;

use aspen::{tree_sitter::Language, Linter};
use lazy_static::lazy_static;

lazy_static! {
    pub static ref RUBY: Language = tree_sitter_ruby::language();
}

fn main() {
    let linter = Linter::new(*RUBY).lints(lints::lints()).comment_str("#");
    linter.run_analysis();
}
