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

use crate::signals::add_signals;
use ansi_term::Colour::RGB;
use chrono::Local;
use config::Config;
use fern::Dispatch;
use panic::setup_panic_handler;
use std::env::var;
use std::thread::current;

mod config;
mod etcd;
mod panic;
mod routing;
mod signals;
mod tql;

type Result<T> = std::result::Result<T, rocket::Error>;

#[rocket::main]
async fn main() -> Result<()> {
    unsafe {
        add_signals();
    }

    // setup logging
    setup_logging();

    // setup config
    Config::new();
    let config = Config::get();

    // setup panic handler
    setup_panic_handler();

    // setup rocket config
    let figment = rocket::Config::figment().merge(("port", config.port.unwrap_or(2314)));
    rocket::custom(figment).launch().await
}

fn setup_logging() {
    let dispatch = Dispatch::new()
        .format(|out, message, record| {
            // If `TSUBASA_DISABLE_COLORS` is enabled as an environment variable
            // then it will not output colours.

            let thread = current();
            let name = match thread.name() {
                Some(s) => s,
                None => "main",
            };

            if let Ok(_) = var("TSUBASA_DISABLE_COLORS") {
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
        .level_for("rocket::launch_", log::LevelFilter::Off)
        .level_for("rocket::shield", log::LevelFilter::Off)
        .level_for("mio::poll", log::LevelFilter::Off)
        .chain(std::io::stdout());

    if let Err(err) = dispatch.apply() {
        panic!("{}", err);
    }
}
