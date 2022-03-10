// 🐇 tsubasa: Microservice to define a schema and execute it in a fast environment.
// Copyright 2022 Noel <cutie@floofy.dev>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

use log::debug;
use once_cell::sync::OnceCell;
use serde::{Deserialize, Serialize};
use std::{env::var, fs};

#[derive(Debug, Serialize, Deserialize)]
pub struct Config {
    pub debug: bool,
    pub http: Option<HttpConfig>,
    pub elastic: ElasticConfig,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct HttpConfig {
    pub port: Option<i32>,
    pub host: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ElasticConfig {
    pub endpoint: String,
    pub username: Option<String>,
    pub password: Option<String>,
    pub indexes: Vec<String>,
}

static INSTANCE: OnceCell<Config> = OnceCell::new();

impl Config {
    pub fn get() -> &'static Config {
        INSTANCE.get().expect("Unable to retrieve global config")
    }

    pub fn create() {
        // Check if we can retrieve the `TSUBASA_CONFIG_PATH` path
        let env_path = var("TSUBASA_CONFIG_PATH");
        let root_path = fs::try_exists("./config.toml");

        if let Ok(path) = env_path {
            debug!("found config path in `TSUBASA_CONFIG_PATH` env variable: '{:?}', now loading from that file...", path);

            let contents =
                fs::read_to_string(path).expect("unable to read from TSUBASA_CONFIG_PATH");

            let result: Config =
                toml::from_str(&contents).expect("cannot serialize `Config` from path");

            INSTANCE.set(result).expect("unable to set global config");
        } else if root_path.is_ok() {
            let path = "./config.toml";
            debug!("found config path in root directory! now loading from that file...");

            let contents = fs::read_to_string(path).expect("unable to read from ./config.toml");

            let result: Config =
                toml::from_str(&contents).expect("cannot serialize `Config` from path");

            INSTANCE.set(result).expect("unable to set global config");
        } else {
            let path = "/app/noel/tsubasa/config.toml";
            debug!("attempting to load from /app/noel/tsubasa/config.toml! now loading from that file...");

            let contents = fs::read_to_string(path).expect("unable to read from ./config.toml");

            let result: Config =
                toml::from_str(&contents).expect("cannot serialize `Config` from path");

            INSTANCE.set(result).expect("unable to set global config");
        }
    }
}

impl Default for HttpConfig {
    fn default() -> Self {
        Self {
            port: Some(23145),
            host: Some("0.0.0.0".to_string()),
        }
    }
}
