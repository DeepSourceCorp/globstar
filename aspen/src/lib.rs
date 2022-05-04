pub mod err;
mod runner;
mod test_utils;
pub use tree_sitter;

use std::any::Any;

use derive_builder::Builder;
use tree_sitter::{Language, Node, Parser, Query, QueryError, Range};

pub type Context = Option<Box<dyn Any>>;

pub struct Linter { lints: Vec<&'static Lint>,
    language: Language,
    comment_str: &'static str,
    context: Context,
}

impl Linter {
    fn __analyze(&self, src: &str) -> Vec<(&'static str, Diagnostic)> {
        let mut parser = Parser::new();
        parser.set_language(self.language).unwrap();

        let syntax_tree = parser.parse(&src, None).unwrap();
        let root_node = syntax_tree.root_node();

        self.lints
            .iter()
            .map(|lint| {
                let diagnostics = (lint.validate)(lint, root_node, &self.context, src.as_bytes());
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
            context: None,
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
pub type Validator = fn(&Lint, Node, &Option<Box<dyn Any>>, &[u8]) -> Vec<Diagnostic>;

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
#[derive(Debug)]
pub struct Diagnostic {
    /// position of offense
    pub at: Range,
    /// context aware offense message
    pub message: String,
}

impl Diagnostic {
    pub fn new<S: AsRef<str>>(at: Range, message: S) -> Self {
        Self {
            at,
            message: message.as_ref().to_owned(),
        }
    }
}

/// Fast failing query-builder
pub fn build_query(language: Language, query_str: &str) -> Query {
    let query = Query::new(language, query_str);
    match query {
        Ok(q) => return q,
        Err(QueryError {
            row,
            column,
            message,
            ..
        }) => {
            log::error!(
                "query builder failed with `{}` on line {}, col {}",
                message,
                row,
                column
            );
            panic!();
        }
    }
}
