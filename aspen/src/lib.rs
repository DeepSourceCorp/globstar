use std::any::Any;

use derive_builder::Builder;
use tree_sitter::{Language, Node, Parser, Query, Range};

pub use tree_sitter;

#[derive(Builder)]
#[builder(pattern = "owned")]
pub struct Linter {
    lints: Vec<&'static Lint>,
    language: Language,
}

impl Linter {
    pub fn analyze(&self, src: &str) -> Vec<Diagnostic> {
        let mut parser = Parser::new();
        parser.set_language(self.language).unwrap();

        let syntax_tree = parser.parse(&src, None).unwrap();
        let root_node = syntax_tree.root_node();

        self.lints
            .iter()
            .map(|lint| (lint.validate)(lint, root_node, None))
            .flatten()
            .collect()
    }
}

/// Describes the rule itself
/// The rule can make use of metadata, the query, anything from the tree and
/// additional context provided by each linter via Box<dyn Any>
pub type Validator = fn(&Lint, Node, Option<Box<dyn Any>>) -> Vec<Diagnostic>;

/// Metadata of a lint
#[derive(Builder)]
#[builder(pattern = "owned")]
pub struct Lint {
    pub name: &'static str,
    pub code: &'static str,
    pub query: &'static Query,
    pub validate: Validator,
}

/// An occurrence of an offense
#[derive(Debug, Builder)]
#[builder(pattern = "owned")]
pub struct Diagnostic {
    /// position of offense
    pub at: Range,
    /// context aware offense message
    pub message: Message,
}

/// Context aware offense message
#[derive(Debug)]
pub struct Message {
    format_string: &'static str,
    binds: Vec<Range>,
}

impl Message {
    pub fn new<V: AsRef<[Range]>>(format_string: &'static str, binds: V) -> Self {
        let binds = binds.as_ref().to_owned();
        let bind_len = binds.len();

        let slots = format_string.matches("${}").count();
        if bind_len != slots {
            eprintln!(
                "attempting to construct malformed message: {} holes, {} binds",
                slots, bind_len
            );
        }
        Self {
            format_string,
            binds,
        }
    }
    pub fn to_string(&self, src: &str) -> String {
        let mut new_string = self.format_string.to_owned();
        for ((idx, _), bind) in self.format_string.match_indices("${}").zip(&self.binds) {
            let replacement = &src[bind.start_byte..bind.end_byte];
            new_string.replace_range(idx..idx + 3, replacement);
        }
        new_string
    }
}
