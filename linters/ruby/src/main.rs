mod lints;

use aspen::{tree_sitter::Language, Linter};
use lazy_static::lazy_static;
use marvin::{config::AnalyzerConfig, Load};

lazy_static! {
    pub static ref RUBY: Language = tree_sitter_ruby::language();
}

fn main() {
    let linter = Linter::new(*RUBY).lints(lints::lints()).comment_str("#");

    let src = r#"
if a = 2
    puts "hi"
else
    puts "hello"
end
    "#;

    let analysis_config = AnalyzerConfig::load();

    linter
        .analyze(src, None)
        .iter()
        .for_each(|d| println!("{:?}", d));
}
