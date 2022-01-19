use std::{any::Any, fs, path::Path};

use crate::{
    err::{AnalysisErr, AspenErr},
    Linter,
};

use marvin::{
    config::AnalyzerConfig,
    err::MarvinErr,
    result::{AnalysisResult, Issue, Location, Position, Span},
    utils::strip_path,
    Load, Store,
};

impl Linter {
    pub fn analysis_runner(&self, ctx: &Option<Box<dyn Any>>) -> Result<(), AspenErr> {
        let config = AnalyzerConfig::load()
            .map_err(MarvinErr::Load)
            .map_err(AspenErr::Marvin)?;
        let (success, failures): (Vec<_>, Vec<_>) = config
            .files
            .into_iter()
            .map(|fq_path| self.analysis_runer_single(fq_path, ctx))
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
            .map_err(AspenErr::Marvin)
    }

    fn analysis_runer_single<P: AsRef<Path>>(
        &self,
        fq_path: P,
        ctx: &Option<Box<dyn Any>>,
    ) -> Result<Vec<Issue>, AnalysisErr> {
        // fully qualified path, use this to read/write
        let fq_path = fq_path.as_ref();
        // stripped path, use this in issue location
        let stripped_path = strip_path(fq_path).map_err(AnalysisErr::Path)?;

        let src = fs::read_to_string(fq_path).map_err(AnalysisErr::Read)?;

        Ok(self
            .analyze(&src, ctx)
            .into_iter()
            .map(|(code, diagnostic)| Issue {
                code,
                message: diagnostic.message.to_string(&src),
                location: Location {
                    path: stripped_path.to_path_buf(),
                    position: Span {
                        begin: Position {
                            line: diagnostic.at.start_point.row,
                            column: diagnostic.at.start_point.column,
                        },
                        end: Position {
                            line: diagnostic.at.end_point.row,
                            column: diagnostic.at.end_point.column,
                        },
                    },
                },
            })
            .collect())
    }
}
