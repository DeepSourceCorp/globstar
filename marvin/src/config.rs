use std::{collections::HashMap, fs, path::PathBuf};

use crate::{err::LoadErr, Load, ANALYSIS_CONFIG_PATH};

use serde::Deserialize;

#[derive(Deserialize, Debug)]
pub struct AnalyzerConfig {
    pub files: Vec<PathBuf>,
    pub exclude_patterns: Option<Vec<String>>,
    pub exclude_files: Vec<PathBuf>,
    pub test_patterns: Option<Vec<String>>,
    pub test_files: Vec<PathBuf>,
    #[serde(default)]
    pub analyzer_meta: AnalyzerMeta,
}

impl Load for AnalyzerConfig {
    fn load() -> Result<Self, LoadErr> {
        let contents = fs::read_to_string(&*ANALYSIS_CONFIG_PATH)?;
        serde_json::from_str(&contents).map_err(LoadErr::Deserialize)
    }
}

#[derive(Default, Deserialize, Debug)]
pub struct AnalyzerMeta {
    pub name: String,
    pub enabled: bool,
    #[serde(default)]
    pub meta: Meta,
}

/// Poorly typed meta, not sure what would work better here
pub type Meta = HashMap<String, String>;

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    #[test]
    fn deserialize_analyzer_config() {
        let config = json!(
            {
                "files": [
                    "file1.py",
                    "file2.py"
                ],
                "exclude_patterns": ["**/migrations/**"],
                "exclude_files": ["core/migrations/0001_init.py"],
                "test_files": ["core/tests/test_core.py"],
                "test_patterns": ["**/tests/test_*.py"],
                "analyzer_meta": {
                    "name": "rust",
                    "enabled": true,
                }
            }
        );
        let serialized_config: AnalyzerConfig = serde_json::from_value(config).unwrap();
        assert_eq!(serialized_config.files.len(), 2);
        assert!(serialized_config
            .files
            .iter()
            .all(|p| p.extension().unwrap() == "py"));
        assert!(serialized_config.analyzer_meta.meta.is_empty());
    }
}
