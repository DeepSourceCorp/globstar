mod lints;

use aspen::{tree_sitter::Language, LinterBuilder};
use lazy_static::lazy_static;

lazy_static! {
    pub static ref RUBY: Language = tree_sitter_ruby::language();
}

fn main() {
    let linter = LinterBuilder::default()
        .language(*RUBY)
        .lints(lints::lints())
        .build()
        .unwrap();

    let src = r#"
        if a = 2
            puts "hello"
        else
            puts "no"
        end

        if a = a
            puts "hello"
        else
            puts "no"
        end
    "#;

    linter
        .analyze(src)
        .iter()
        .for_each(|d| println!("{}", d.message.to_string(&src)));
}
