use std::{fs, path::Path};

use crate::{
    err::{AnalysisErr, GlobstarErr},
    Lint, Linter, Occurrence,
};

use marvin::{
    config::AnalyzerConfig,
    err::MarvinErr,
    result::{AnalysisResult, Issue, Location, Position, Span},
    utils::strip_path,
    Load, Store,
};
use regex::RegexSet;

impl Linter {
    /// The whole point.
    ///
    /// `run_analysis` does the following:
    /// - read environment variables required to interface with `marvin`
    /// - load `analysis_config.json` from `$ANALYSIS_CONFIG_PATH`
    /// - build an ignore ruleset to avoid ignored files
    /// - build scope and injection data
    /// - walk through the files present in `$CODE_PATH`
    /// - store `analysis_results.json` to `$ANALYSIS_RESULT_PATH`
    pub fn run_analysis(&self) -> Result<(), GlobstarErr> {
        let config = AnalyzerConfig::load()
            .map_err(MarvinErr::Load)
            .map_err(GlobstarErr::Marvin)?;

        let ignore_set = RegexSet::new(&self.ignores).map_err(GlobstarErr::Ignore)?;

        let (success, _failures): (Vec<_>, Vec<_>) = config
            .files
            .into_iter()
            .filter(|file| !ignore_set.is_match(&file.to_string_lossy()))
            .filter(|file| matches!(file.extension(), Some(ext) if ext == self.extension.as_str()))
            .map(|fq_path| self.analysis_runner_single(fq_path))
            .partition(Result::is_ok);
        let success = success.into_iter().map(Result::unwrap).flatten().collect();
        // let failures = failures.into_iter().map(Result::unwrap_err).collect();
        let result = AnalysisResult {
            issues: success,
            ..Default::default()
        };
        result
            .store()
            .map_err(MarvinErr::Store)
            .map_err(GlobstarErr::Marvin)
    }

    fn analysis_runner_single<P: AsRef<Path>>(
        &self,
        fq_path: P,
    ) -> Result<Vec<Issue>, AnalysisErr> {
        // fully qualified path, use this to read/write
        let fq_path = fq_path.as_ref();
        // stripped path, use this in issue location
        let stripped_path = strip_path(fq_path).map_err(AnalysisErr::Path)?;

        let src = fs::read_to_string(fq_path).map_err(AnalysisErr::Read)?;

        Ok(self
            .__analyze(&src)
            .into_iter()
            .map(
                |Occurrence {
                     lint: Lint { code, .. },
                     at,
                     message,
                     ..
                 }| Issue {
                    code,
                    message,
                    location: Location {
                        path: stripped_path.to_path_buf(),
                        position: Span {
                            begin: Position {
                                line: at.start_point.row + 1,
                                column: at.start_point.column + 1,
                            },
                            end: Position {
                                line: at.end_point.row + 1,
                                column: at.end_point.column + 1,
                            },
                        },
                    },
                },
            )
            .collect())
    }
}
