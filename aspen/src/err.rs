use std::io;

use marvin::err::MarvinErr;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum AspenErr {
    #[error("marvin error: {0}")]
    Marvin(#[from] MarvinErr),
    #[error("runner error: {0}")]
    Runner(RunnerErr),
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
