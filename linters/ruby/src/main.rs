mod lints;

use aspen::{tree_sitter::Language, Linter};
use lazy_static::lazy_static;

lazy_static! {
    pub static ref RUBY: Language = tree_sitter_ruby::language();
}

fn main() {
    let linter = Linter::new(*RUBY).lints(lints::lints()).comment_str("#");

    let src = r#"
    system <<-BASH.strip!
        abc --def | ghi > jkl
    BASH
    "#;

    linter
        .analyze(src)
        .iter()
        .for_each(|d| println!("{}", d.message.to_string(&src)));
}
