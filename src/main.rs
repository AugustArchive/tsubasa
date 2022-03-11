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

#![feature(path_try_exists)]

#[macro_use]
extern crate rocket;

#[macro_use]
extern crate log;

use crate::elastic::Elasticsearch;
use crate::routing::*;
use crate::signals::add_signals;
use ansi_term::Colour::RGB;
use chrono::Local;
use config::{Config, HttpConfig};
use fern::Dispatch;
use rocket::routes;
use rocket_prometheus::PrometheusMetrics;
use std::env::var;
use std::thread::current;

mod config;
mod elastic;
mod routing;
mod signals;

type Result<T> = std::result::Result<T, rocket::Error>;

#[rocket::main]
async fn main() -> Result<()> {
    unsafe {
        add_signals();
    }

    Config::create();
    let config = Config::get();

    // setup logging
    setup_logging(config);

    // setup elasticsearch
    info!("setting up elasticsearch...");

    Elasticsearch::create();
    let elastic = Elasticsearch::get();
    elastic
        .test_connection()
        .await
        .expect("Unable to test Elasticsearch connection");

    let default_http = HttpConfig::default();
    let http_config = config.http.as_ref().unwrap_or(&default_http);

    // setup rocket config
    let figment = rocket::Config::figment()
        .merge(("port", http_config.port.unwrap_or(23145)))
        .merge((
            "host",
            http_config.host.as_ref().unwrap_or(&"0.0.0.0".to_string()),
        ));

    let metrics = PrometheusMetrics::with_default_registry();
    let server = rocket::custom(figment)
        .attach(metrics.clone())
        .mount("/", routes![hello, health])
        .mount("/index", routes![index_get, index_fetch, index_search])
        .mount("/metrics", metrics);

    server.launch().await
}

fn setup_logging(config: &'static Config) {
    let is_debug = if let Some(debug) = config.debug {
        debug
    } else {
        false
    };

    let dispatch = Dispatch::new()
        .format(|out, message, record| {
            // If `TSUBASA_DISABLE_COLORS` is enabled as an environment variable
            // then it will not output colours.

            let thread = current();
            let name = thread.name().unwrap_or("main");

            if var("TSUBASA_DISABLE_COLORS").is_ok() {
                out.finish(format_args!(
                    "{} [{} <{}>] {} :: {}",
                    Local::now().format("[%B %d, %G | %H:%M:%S %p]"),
                    record.target(),
                    name,
                    record.level(),
                    message
                ));
            } else {
                let now = Local::now().format("[%B %d, %G | %H:%M:%S %p]");
                let level_color = match record.level() {
                    log::Level::Error => RGB(153, 75, 104),
                    log::Level::Debug => RGB(163, 182, 138),
                    log::Level::Info => RGB(178, 157, 243),
                    log::Level::Trace => RGB(163, 182, 138),
                    log::Level::Warn => RGB(243, 243, 134),
                };

                out.finish(format_args!(
                    "{} {}{}{} {} :: {}",
                    RGB(134, 134, 134).paint(format!("{}", now)),
                    RGB(178, 157, 243).paint(format!("[{} ", record.target())),
                    RGB(255, 105, 189).paint(format!("<{}>", name)),
                    RGB(178, 157, 243).paint("]"),
                    level_color.paint(format!("{}", record.level())),
                    message
                ));
            }
        })
        // turn off rocket-related stuff
        .level_for("_", log::LevelFilter::Off)
        // .level_for("rocket::launch_", log::LevelFilter::Off)
        .level_for("rocket::shield", log::LevelFilter::Off)
        .level_for("mio::poll", log::LevelFilter::Off)
        .level_for("want", log::LevelFilter::Off)
        .level_for("tokio_util::codec::framed_impl", log::LevelFilter::Off)
        .level(if is_debug {
            log::LevelFilter::Debug
        } else {
            log::LevelFilter::Info
        })
        .chain(std::io::stdout());

    if let Err(err) = dispatch.apply() {
        panic!("{}", err);
    }
}
