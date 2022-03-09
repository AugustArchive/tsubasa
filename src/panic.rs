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

use log::{debug, error};
use std::{panic, thread::current};

pub fn setup_panic_handler() {
    debug!("setting up custom panic handler...");
    panic::set_hook(Box::new(|info| {
        let curr = current();
        let name = curr.name().unwrap_or("main");
        let msg = match info.payload().downcast_ref::<&'static str>() {
            Some(s) => *s,
            None => match info.payload().downcast_ref::<String>() {
                Some(s) => &**s,
                None => "Box<?>",
            },
        };

        let backtrace = backtrace::Backtrace::new();
        match info.location() {
            Some(location) => {
                error!(
                    "thread '{}' panicked at '{}' ({}:{})\n{:?}",
                    name,
                    msg,
                    location.file(),
                    location.line(),
                    backtrace
                );
            }

            None => {
                error!("thread '{}' panicked at '{}'\n{:?}", name, msg, backtrace);
            }
        }
    }));
}
