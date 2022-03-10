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

use crate::elastic::Elasticsearch;
use rocket::serde::json::{serde_json::json, Json, Value};
use serde::Deserialize;

#[derive(Deserialize, Clone, Copy)]
pub enum MatchType {
    MatchAll,
    Fuzzy,
}

#[derive(Deserialize, Clone)]
pub struct IndexSearchBody {
    pub match_type: MatchType,
    pub data: Value,
}

// ~ NORMAL ENDPOINTS ~

#[get("/")]
pub fn hello() -> Value {
    json!({
        "success": true,
        "message": "hello world!"
    })
}

#[get("/health")]
pub fn health() -> &'static str {
    "OK"
}

// ~ INDEX INFORMATION ~
#[get("/")]
pub fn index_get() -> Value {
    json!({
        "success": true,
        "message": "Welcome to the indexes API!"
    })
}

#[get("/<index>")]
pub async fn index_fetch(index: &str) -> Value {
    let elastic = Elasticsearch::get();
    let exists = elastic.index_exists(index).await;
    let (count, deleted) = elastic.index_doc_stats(index).await;

    json!({ "success": true, "data": { "exists": exists, "count": count, "deleted": deleted } })
}

#[post("/<index>/search", data = "<body>")]
pub async fn index_search(index: &str, body: Json<IndexSearchBody>) -> Value {
    let elastic = Elasticsearch::get();
    elastic
        .index_search(index, body.match_type, body.data.clone())
        .await
}
