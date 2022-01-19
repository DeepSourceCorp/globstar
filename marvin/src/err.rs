use std::io;

use thiserror::Error;

#[derive(Error, Debug)]
pub enum MarvinErr {
    #[error("load error: {0}")]
    Load(LoadErr),
    #[error("store error: {0}")]
    Store(StoreErr),
}

#[derive(Error, Debug)]
pub enum LoadErr {
    #[error("read error: {0}")]
    Read(#[from] io::Error),
    #[error("deserialization error: {0}")]
    Deserialize(#[from] serde_json::Error),
}

#[derive(Error, Debug)]
pub enum StoreErr {
    #[error("write error: {0}")]
    Write(#[from] io::Error),
    #[error("serialization error: {0}")]
    Serialize(#[from] serde_json::Error),
}
