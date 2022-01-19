pub mod config;
pub mod err;
pub mod result;
pub mod utils;

use std::env;

// set global static path variables from environment at run time
// if variables are not present, use sane defaults
environment!(CODE_PATH);
environment!(ANALYSIS_CONFIG_PATH);
environment!(ANALYSIS_RESULT_PATH);
environment!(AUTOFIX_CONFIG_PATH);
environment!(AUTOFIX_RESULT_PATH);

// helper macro to read from environment at run time and initialize a static global
#[macro_export]
macro_rules! environment {
    ($x:ident) => {
        ::lazy_static::lazy_static! {
            pub(crate) static ref $x: String = env::var(stringify!($x)).unwrap_or(defaults::$x.into());
        }
    };
}

// paths to default to if marvin-rust can't find these in environment
mod defaults {
    pub(crate) static CODE_PATH: &'static str = "/code";
    pub(crate) static ANALYSIS_CONFIG_PATH: &'static str = "/toolbox/analysis_config.json";
    pub(crate) static ANALYSIS_RESULT_PATH: &'static str = "/toolbox/analysis_results.json";
    pub(crate) static AUTOFIX_CONFIG_PATH: &'static str = "/toolbox/autofix_config.json";
    pub(crate) static AUTOFIX_RESULT_PATH: &'static str = "/toolbox/autofix_results.json";
}

pub trait Load {
    fn load() -> Result<Self, err::LoadErr>
    where
        Self: Sized;
}

pub trait Store {
    fn store(&self) -> Result<(), err::StoreErr>;
}
