pub mod err;
mod runner;
mod test_utils;
pub use tree_sitter;

use std::any::Any;

use derive_builder::Builder;
use tree_sitter::{Language, Node, Parser, Query, Range};

pub struct Linter {
    lints: Vec<&'static Lint>,
    language: Language,
    comment_str: &'static str,
}

impl Linter {
    /// Run analysis on given source code, producing a list of diagnostics
    pub fn analyze(
        &self,
        src: &str,
        ctx: &Option<Box<dyn Any>>,
    ) -> Vec<(&'static str, Diagnostic)> {
        let mut parser = Parser::new();
        parser.set_language(self.language).unwrap();

        let syntax_tree = parser.parse(&src, None).unwrap();
        let root_node = syntax_tree.root_node();

        self.lints
            .iter()
            .map(|lint| {
                let diagnostics = (lint.validate)(lint, root_node, ctx);
                diagnostics.into_iter().map(|d| (lint.code, d))
            })
            .flatten()
            .collect()
    }

    /// Create a new Linter instance of a language
    pub fn new(language: Language) -> Self {
        Self {
            lints: vec![],
            language,
            comment_str: "//",
        }
    }

    /// Set this Linter's language
    pub fn language(mut self, language: Language) -> Self {
        self.language = language;
        self
    }

    /// Add a lint to this Linter
    pub fn lint(mut self, lint: &'static Lint) -> Self {
        self.lints.push(lint);
        self
    }

    /// Set a list of lints accepted by this linter
    pub fn lints(mut self, lints: Vec<&'static Lint>) -> Self {
        self.lints = lints;
        self
    }

    /// Set the comment str accepted by this language, this is used
    /// in annotated tests
    pub fn comment_str(mut self, comment_str: &'static str) -> Self {
        self.comment_str = comment_str;
        self
    }
}

/// Describes the rule itself
/// The rule can make use of metadata, the query, anything from the tree and
/// additional context provided by each linter via Box<dyn Any>
pub type Validator = fn(&Lint, Node, &Option<Box<dyn Any>>) -> Vec<Diagnostic>;

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
