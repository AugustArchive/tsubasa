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

// use rocket::{
//     response::{content::Json, Responder, Response, Result as ResponseResult},
//     serde::json::Value,
// };
use serde::{Deserialize, Serialize};

/// Represents an error that might occur within request execution
/// of Tsubasa.
#[derive(Deserialize, Serialize)]
pub enum Error {
    /// [**schema execution**] - This error occurs if a field
    /// doesn't exist in the current schema.
    UnknownField { field: String },

    /// [**schema parsing**] - This error occurs if the data type
    /// was not registered within a project's schema.
    UnknownDataType { data_type: String },
}

// #[rocket::async_trait]
// impl<'r> Responder<'r, 'static> for std::result::Result<Json<Value>, Error> {
//     fn respond_to(self, _: &'r Request<'_>) -> ResponseResult<'static> {}
// }
