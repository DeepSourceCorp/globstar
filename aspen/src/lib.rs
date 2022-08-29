// exports
pub mod err;
pub use {
    context::Context,
    traits::{IsMatch, MapCapture},
    tree_sitter,
};

mod context;
mod runner;
#[cfg(feature = "testing")]
mod test_utils;
mod traits;

use std::fmt;

use crate::err::InjectionErr;

use cedar::ResolutionMethod;
use context::InjectedTree;
use tree_sitter::{Language, Node, Parser, Query, QueryCursor, QueryError, Range};

pub struct Linter {
    validators: Vec<ValidatorFn>,
    language: Language,
    comment_str: &'static str,
    scopes: Option<&'static str>,
    injections: Vec<Injection>,
    ignores: Vec<&'static str>,
    extension: &'static str,
}

pub struct Injection {
    query: Query,
    language: Language,
}

impl Injection {
    pub fn new(query: &'static str, language: Language) -> Result<Self, InjectionErr> {
        let query = Query::new(language, query).map_err(InjectionErr::Query)?;
        if !query
            .capture_names()
            .iter()
            .any(|name| name != "injection.content")
        {
            Err(InjectionErr::MissingCapture)
        } else {
            Ok(Self { query, language })
        }
    }
}

impl Linter {
    fn __analyze(&self, src: &str) -> Vec<Occurrence> {
        let mut parser = Parser::new();
        parser.set_language(self.language).unwrap();

        let syntax_tree = parser.parse(&src, None).unwrap();
        let root_node = syntax_tree.root_node();

        // TODO; handle AspenErr here
        let context = self.build_context(root_node, src).expect("query error");

        self.validators
            .iter()
            .map(|v| v(root_node, &context, src.as_bytes()))
            .flatten()
            .collect()
    }

    pub fn build_context<'a>(
        &self,
        root_node: Node<'a>,
        src: &'a str,
    ) -> Result<Option<Context<'a>>, QueryError> {
        let root_scope = self
            .scopes
            .map(|scope_query| {
                Query::new(self.language, scope_query)
                    .map(|query| ResolutionMethod::Generic.build_scope(&query, root_node, src))
            })
            .transpose()?;

        let injected_trees = self
            .injections
            .iter()
            .filter_map(|Injection { query, language }| {
                // run the injection query with the syntax tree formed from the original language
                let mut injection_parser = Parser::new();
                injection_parser.set_language(*language).unwrap();

                // infallible, this is checked for in `Injection::new`
                let content_capture = query.capture_index_for_name("injection.content").unwrap();

                Some(
                    QueryCursor::new()
                        .matches(&query, root_node, src.as_bytes())
                        .flat_map(|m| m.captures)
                        .filter(|c| c.index == content_capture)
                        .filter_map(|capture| {
                            let original_range = capture.node.range();
                            let (start, end) = (original_range.start_byte, original_range.end_byte);
                            injection_parser.parse(&src[start..end], None).map(|tree| {
                                InjectedTree {
                                    tree,
                                    original_range,
                                }
                            })
                        })
                        .collect::<Vec<_>>(),
                )
            })
            .flatten()
            .collect::<Vec<_>>();

        Ok(root_scope.map(|root_scope| Context {
            root_scope,
            injected_trees,
        }))
    }

    /// Create a new Linter instance of a language
    pub fn new(language: Language) -> Self {
        Self {
            validators: vec![],
            language,
            comment_str: "//",
            scopes: None,
            injections: vec![],
            ignores: vec![],
            extension: "",
        }
    }

    /// Set this Linter's language
    pub fn language(mut self, language: Language) -> Self {
        self.language = language;
        self
    }

    /// Add a lint to this Linter
    pub fn validator(mut self, validator: ValidatorFn) -> Self {
        self.validators.push(validator);
        self
    }

    /// Set a list of lints accepted by this linter
    pub fn validators(mut self, validators: &[ValidatorFn]) -> Self {
        self.validators = validators.to_vec();
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

    /// Language injection queries
    pub fn injection(mut self, injection: Injection) -> Self {
        self.injections.push(injection);
        self
    }

    /// Add an ignore pattern
    pub fn ignore(mut self, regex: &'static str) -> Self {
        self.ignores.push(regex);
        self
    }

    /// Set the list of ignores
    pub fn ignores(mut self, regex_set: &[&'static str]) -> Self {
        self.ignores = regex_set.to_vec();
        self
    }

    /// Set the file extension
    pub fn extension(mut self, extension: &'static str) -> Self {
        self.extension = extension;
        self
    }
}

/// Describes the rule itself
pub type ValidatorFn = for<'a> fn(Node, &Option<Context<'a>>, &[u8]) -> Vec<Occurrence>;

#[derive(Debug)]
pub struct Occurrence {
    pub name: &'static str,
    pub code: &'static str,
    pub diagnostic: Diagnostic,
}

impl fmt::Display for Occurrence {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}, {}, {}", self.name, self.code, self.diagnostic)
    }
}

pub struct Lint {
    pub name: &'static str,
    pub code: &'static str,
}

impl Lint {
    pub fn raise<S: AsRef<str>>(&self, at: Range, message: S) -> Occurrence {
        Occurrence {
            name: self.name,
            code: self.code,
            diagnostic: Diagnostic::new(at, message),
        }
    }
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

impl fmt::Display for Diagnostic {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "{}..{}: {}",
            self.at.start_point, self.at.end_point, self.message
        )
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
