use std::io;

use marvin::err::MarvinErr;
use regex::Error as RegexError;
use thiserror::Error;
use tree_sitter::QueryError as TSQueryError;

/// Errors that can occur when running a `globstar` linter,
/// this type is a wrapper around various errors raised by `globstar`.
#[derive(Error, Debug)]
#[non_exhaustive]
pub enum GlobstarErr {
    /// Wrapper around [`MarvinErr`](MarvinErr)
    #[error("marvin error: {0}")]
    Marvin(#[from] MarvinErr),

    /// Wrapper around [`RunnerErr`](RunnerErr)
    #[error("runner error: {0}")]
    Runner(RunnerErr),

    /// Wrapper around [`tree_sitter::QueryError`](TSQueryError)
    #[error("query error: {0}")]
    Query(#[from] TSQueryError),

    /// Wrapper around [`regex::Error`](RegexError)
    #[error("regex error: {0}")]
    Ignore(#[from] RegexError),

    /// Wrapper around [`InjectionErr`](InjectionErr)
    #[error("injection errer: {0}")]
    Injection(#[from] InjectionErr),
}

/// Raised when [`Injection`s](crate::Injection) are created.
///
/// `InjectionErr` covers two scenarios:
///
/// - The injection query is not a valid tree-sitter query
/// - The injection query does not contain an `injection.content`
/// capture group
#[derive(Error, Debug)]
pub enum InjectionErr {
    /// Raised when the injection query is not a valid
    /// tree-sitter query
    #[error("query error: {0}")]
    Query(#[from] TSQueryError),

    /// Raised when the injection query does not contain an
    /// `injection.content` capture group
    #[error("missing capture `injection.content`")]
    MissingCapture,
}

/// Errors that can occur during construction or execution of a linter
#[derive(Error, Debug)]
#[non_exhaustive]
pub enum RunnerErr {
    /// Raised on crucial analysis errors
    #[error("analysis error: {0}")]
    Analysis(AnalysisErr),
}

/// Errors that can occur during an analysis run
#[derive(Error, Debug)]
#[non_exhaustive]
pub enum AnalysisErr {
    /// Raised on IO errors while reading the code paths from disk
    #[error("read error: {0}")]
    Read(#[from] io::Error),

    /// Raised when the runner is unable to strip `$CODE_PATH` from a file path
    #[error("non-conformant path: {0}")]
    Path(std::path::StripPrefixError),
}
