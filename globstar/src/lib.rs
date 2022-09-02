#![deny(missing_docs)]
//! `globstar` is a generic analysis framework.

/// Error types for the globstar crate.
pub mod err;

/// Utility traits to write lint rules.
pub mod traits;

/// re-exported tree-sitter.
pub use tree_sitter;

mod context;
mod runner;
#[cfg(feature = "testing")]
mod test_utils;

use std::fmt;

use context::{Context, InjectedTree};
use err::InjectionErr;
use scope_resolution::ResolutionMethod;
use tree_sitter::{Language, Node, Parser, Query, QueryCursor, QueryError, Range};

/// The entrypoint into running an analysis.
///
/// A `Linter` contains information about the various rules,
/// scope queries, language injections etc. It is the entrypoint
/// to running an analysis.
pub struct Linter {
    validators: Vec<ValidatorFn>,
    language: Language,
    comment_str: &'static str,
    scopes: Option<&'static str>,
    injections: Vec<Injection>,
    ignores: Vec<&'static str>,
    extension: &'static str,
}

/// An "injection" defines the rules to parse a language within a language.
///
/// Several languages "nest" one language within another. Globstar allows
/// you to parse and analyze such nested languages in addition to analyzing
/// the outer program. For example, an Ansible progam, largely contains YAML
/// code, but certain keys may contain jinja or bash:
///
/// ```yaml
/// - name: use shell generator
///   ansible.builtin.shell: ls foo{.txt,.xml}
///                        ## ^^^^^^^^^^^^^^^^^ bash code
///   changed_when: false
/// ```
///
/// The key `ansible.builtin.shell` takes a shell script as its value. In order
/// to treat the value as Bash, and not a YAML string, you may write a query to
/// capture and parse the value as Bash:
///
/// ```
/// # use globstar::{Linter, Injection};
/// # fn main() {
/// let yaml = tree_sitter_yaml::language();
/// let shell_query = r#"
///     ((block_mapping_pair
///       key: (_) @key
///       value:
///       (flow_node
///         (plain_scalar
///           (string_scalar) @injection.content)))
///     (#match? @key "^(ansible\\.builtin\\.)?shell")
///     (#set! injection.language "bash"))
/// "#;
/// let bash_injection = Injection::new(shell_query, yaml).unwrap();
/// let ansible_analyzer = Linter::new(yaml)
///     .injection(bash_injection);
/// # }
/// ```
pub struct Injection {
    query: Query,
    language: Language,
}

impl Injection {
    /// Create a new `Injection` with a given query for a nested language. The `query`
    /// **must** include a capture called `injection.content`. The result of the capture
    /// is parsed as `language`.
    ///
    /// Note: `query` is a tree-sitter query in the outer language, `language`
    /// is the nested language, which is then used to parse the result of the
    /// query.
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

    fn build_context<'a>(
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

    /// Add an ignore pattern, files conforming to this pattern
    /// are not processed.
    ///
    /// Example:
    ///
    /// ```
    /// # use globstar::Linter;
    /// let linter = Linter::new(tree_sitter_bash::language())
    ///     .ignore(r"Cargo\.toml");
    /// ```
    pub fn ignore(mut self, regex: &'static str) -> Self {
        self.ignores.push(regex);
        self
    }

    /// Set a list patterns for files to ignore.
    ///
    /// Example:
    pub fn ignores(mut self, regex_set: &[&'static str]) -> Self {
        self.ignores = regex_set.to_vec();
        self
    }

    /// Set the file extension, e.g.: `"yml"`
    pub fn extension(mut self, extension: &'static str) -> Self {
        self.extension = extension;
        self
    }
}

/// Analysis rule logic goes in here
///
/// Every `ValidatorFn` recives the `root_node` for a syntax tree, along with
/// (optionally) scope and injection data, and source file (as bytes).
pub type ValidatorFn = for<'a> fn(Node, &Option<Context<'a>>, &[u8]) -> Vec<Occurrence>;

/// Metadata about an antipattern that the `Linter` raises.
///
/// TODO: examples
#[derive(Copy, Clone, Debug)]
#[non_exhaustive]
pub struct Lint {
    /// The name of this lint, for e.g.: `UNUSED_VARIABLES`
    pub name: &'static str,
    /// A unique code to identify this lint, for e.g.: `RUST-001`
    pub code: &'static str,
}

impl Lint {
    /// To produce an [`Occurrence`](Occurrence) from a `Lint`, use `Lint::raise`,
    /// along with a position and a context aware message.
    pub fn raise<S: AsRef<str>>(&self, at: Range, message: S) -> Occurrence {
        Occurrence {
            lint: *self,
            at,
            message: message.as_ref().to_owned(),
        }
    }
}

/// An instance of a `Lint` is an `Occurrence`.
#[derive(Debug)]
#[non_exhaustive]
pub struct Occurrence {
    /// The lint whose occurrence this refers to.
    pub lint: Lint,
    /// The position where this occurrence is present at.
    pub at: Range,
    /// A context aware message describing this occurrence.
    pub message: String,
}

impl fmt::Display for Occurrence {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "{}, {}, {}: {}..{}",
            self.lint.name, self.lint.code, self.message, self.at.start_point, self.at.end_point
        )
    }
}
