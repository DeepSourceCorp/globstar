use std::io;

use marvin::err::MarvinErr;
use regex::Error as RegexError;
use thiserror::Error;
use tree_sitter::QueryError as TSQueryError;

#[derive(Error, Debug)]
pub enum AspenErr {
    #[error("marvin error: {0}")]
    Marvin(#[from] MarvinErr),
    #[error("runner error: {0}")]
    Runner(RunnerErr),
    #[error("query error: {0}")]
    Query(#[from] TSQueryError),
    #[error("regex error: {0}")]
    Ignore(#[from] RegexError),
    #[error("injection errer: {0}")]
    Injection(#[from] InjectionErr),
}

#[derive(Error, Debug)]
pub enum InjectionErr {
    #[error("query error: {0}")]
    Query(#[from] TSQueryError),
    #[error("missing capture `injection.content`")]
    MissingCapture,
}

#[derive(Error, Debug)]
pub enum RunnerErr {
    #[error("analysis error: {0}")]
    Analysis(AnalysisErr),
}

#[derive(Error, Debug)]
pub enum AnalysisErr {
    #[error("read error: {0}")]
    Read(#[from] io::Error),
    #[error("non-conformant path: {0}")]
    Path(std::path::StripPrefixError),
}
