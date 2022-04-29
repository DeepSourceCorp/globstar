pub mod err;
mod runner;
mod test_utils;
pub use tree_sitter;

use std::{cell::RefCell, rc::Rc};

use cedar::{LocalScope, ResolutionMethod, ScopeStack};
use derive_builder::Builder;
use tree_sitter::{Language, Node, Parser, Query, QueryError, Range};

pub struct Context<'a> {
    root_scope: Rc<RefCell<LocalScope<'a>>>,
}

impl<'a> Context<'a> {
    pub fn root_scope(&self) -> Rc<RefCell<LocalScope<'a>>> {
        Rc::clone(&self.root_scope)
    }
    pub fn scope_by_range(&self, range: &Range) -> Option<Rc<RefCell<LocalScope<'a>>>> {
        cedar::scope_by_range(self.root_scope(), range)
    }
    pub fn scope_of(&self, node: Node) -> Option<Rc<RefCell<LocalScope<'a>>>> {
        self.scope_by_range(&node.range())
    }
    pub fn scope_stack_by_range(&self, range: &Range) -> Option<ScopeStack<'a>> {
        Some(cedar::scope_stack(self.scope_by_range(range)?))
    }
    pub fn scope_stack_of(&self, node: Node) -> Option<ScopeStack<'a>> {
        self.scope_stack_by_range(&node.range())
    }
}

pub struct Linter {
    lints: Vec<&'static Lint>,
    language: Language,
    comment_str: &'static str,
    scopes: Option<&'static str>,
}

impl Linter {
    fn __analyze(&self, src: &str) -> Vec<(&'static str, Diagnostic)> {
        let mut parser = Parser::new();
        parser.set_language(self.language).unwrap();

        let syntax_tree = parser.parse(&src, None).unwrap();
        let root_node = syntax_tree.root_node();
        let context = self
            .scopes
            .and_then(|scope_query| Query::new(self.language, scope_query).ok())
            .as_ref()
            .map(|query| ResolutionMethod::Generic.build_scope(query, root_node, src))
            .map(|root_scope| Context { root_scope });

        self.lints
            .iter()
            .map(|lint| {
                let diagnostics = (lint.validate)(lint, root_node, &context, src.as_bytes());
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
            scopes: None,
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

    /// Scope resolution queries
    pub fn scopes(mut self, queries: &'static str) -> Self {
        self.scopes = Some(queries);
        self
    }
}

/// Describes the rule itself
/// The rule can make use of metadata, the query, anything from the tree and
/// additional context provided by each linter via Box<dyn Any>
pub type Validator = for<'a> fn(&Lint, Node, &Option<Context<'a>>, &[u8]) -> Vec<Diagnostic>;

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
