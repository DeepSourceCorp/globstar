use std::{fs, path::PathBuf};

use crate::{err::StoreErr, Store, ANALYSIS_RESULT_PATH};

use serde::Serialize;

#[derive(Default, Debug, Serialize)]
pub struct AnalysisResult {
    pub issues: Vec<Issue>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub metrics: Vec<Metric>,
    pub is_passed: bool,
    pub errors: Vec<Error>,
}

impl Store for AnalysisResult {
    fn store(&self) -> Result<(), StoreErr> {
        // technically this cant error out, but w/e
        let contents = serde_json::to_string(&self).map_err(StoreErr::Serialize)?;
        fs::write(&*ANALYSIS_RESULT_PATH, contents.as_bytes()).map_err(StoreErr::Write)
    }
}

#[derive(Debug, Serialize)]
pub struct Issue {
    #[serde(rename = "issue_code")]
    pub code: &'static str,
    #[serde(rename = "issue_text")]
    pub message: String,
    pub location: Location,
}

#[derive(Debug, Serialize)]
pub struct Location {
    pub path: PathBuf,
    pub position: Span,
}

#[derive(Debug, PartialEq, Serialize)]
pub struct Span {
    pub begin: Position,
    pub end: Position,
}

#[derive(Debug, PartialEq, Serialize)]
pub struct Position {
    pub line: usize,
    pub column: usize,
}

#[derive(Debug, Serialize)]
pub struct Metric {
    #[serde(rename = "metric_code")]
    pub code: String,
    pub namespaces: Vec<Namespace>,
}

#[derive(Debug, Serialize)]
pub struct Namespace {
    pub key: String,
    pub value: u64,
}

#[derive(Debug, Serialize)]
pub struct Error {
    pub hmessage: String,
    pub level: u64,
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    #[test]
    fn serialize_analysis_result() {
        let expected = json!(
            {
                "issues": [
                {
                    "issue_code": "SCC-U1000",
                    "issue_text": "func warnings1 is unused",
                    "location": {
                        "path": "overlapping_edits.go",
                        "position": {
                            "begin": {
                                "line": 3,
                                "column": 6
                            },
                            "end": {
                                "line": 3,
                                "column": 6
                            }
                        }
                    }
                }
                ],
                "metrics": [
                {
                    "metric_code": "DCV",
                    "namespaces": [
                    {
                        "key": "Go",
                        "value": 0
                    }
                    ]
                }
                ],
                "is_passed": true,
                "errors": [
                {
                    "hmessage": "Could not install dependencies",
                    "level": 1
                }
                ],
            }
        );
        let observed = AnalysisResult {
            issues: vec![Issue {
                code: "SCC-U1000",
                message: "func warnings1 is unused".into(),
                location: Location {
                    path: "overlapping_edits.go".into(),
                    position: Span {
                        begin: Position { line: 3, column: 6 },
                        end: Position { line: 3, column: 6 },
                    },
                },
            }],
            metrics: vec![Metric {
                code: "DCV".into(),
                namespaces: vec![Namespace {
                    key: "Go".into(),
                    value: 0,
                }],
            }],
            is_passed: true,
            errors: vec![Error {
                hmessage: "Could not install dependencies".into(),
                level: 1,
            }],
        };

        assert_eq!(serde_json::to_value(observed).unwrap(), expected);
    }
}
