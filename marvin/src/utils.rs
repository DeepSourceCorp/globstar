use crate::CODE_PATH;
use std::path::{Path, StripPrefixError};

pub fn strip_path(p: &Path) -> Result<&Path, StripPrefixError> {
    p.strip_prefix(&*CODE_PATH)
}
