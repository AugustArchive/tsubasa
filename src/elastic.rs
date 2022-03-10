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

use crate::{config::Config, routing::MatchType};
use elasticsearch::{
    auth::Credentials,
    cluster::ClusterHealthParts,
    http::{
        headers::HeaderMap,
        transport::{SingleNodeConnectionPool, TransportBuilder},
        Method, Url,
    },
    Elasticsearch as ES, SearchParts,
};
use once_cell::sync::OnceCell;
use serde_json::{json, Value};

#[derive(Debug, Clone)]
pub struct Elasticsearch {
    client: ES,
}

static INSTANCE: OnceCell<Elasticsearch> = OnceCell::new();

impl Elasticsearch {
    pub fn get() -> &'static Elasticsearch {
        INSTANCE
            .get()
            .expect("Cannot retrieve global Elastic client")
    }

    pub fn create() {
        info!("now connecting to elasticsearch...");

        let config = Config::get();
        let uri = Url::parse(&config.elastic.endpoint).expect("Unable to parse endpoint URI.");
        let pool = SingleNodeConnectionPool::new(uri);
        let mut transport_builder = TransportBuilder::new(pool);

        let auth = if let Some(username) = config.elastic.username.as_ref() {
            let password = config
                .elastic
                .password
                .as_ref()
                .expect("Missing `password` field if `username` is populated.");

            Some(Credentials::Basic(
                username.to_string(),
                password.to_string(),
            ))
        } else {
            None
        };

        if let Some(credentials) = auth {
            transport_builder = transport_builder.auth(credentials);
        }

        let transport = transport_builder
            .build()
            .expect("Unable to build Elastic transport");

        let client = ES::new(transport);
        let es = Elasticsearch { client };

        INSTANCE
            .set(es)
            .expect("Unable to set global Elastic instance");
    }

    pub async fn test_connection(&self) -> Result<(), &'static str> {
        info!("testing elastic client connection...");

        let api = self.client.cluster();
        let res = api
            .health(ClusterHealthParts::None)
            .send()
            .await
            .expect("Unable to request to Elastic");

        let code = res.status_code();
        let body = res
            .json::<Value>()
            .await
            .expect("Unable to deserialise payload");

        let cluster_name = body["cluster_name"]
            .as_str()
            .expect("Unable to retrieve cluster name");

        let status = body["status"]
            .as_str()
            .expect("Unable to retrieve cluster status");

        if status == "red" {
            return Err(
                "Cluster is not healthy, please restart Tsubasa once the ES cluster is available!",
            );
        }

        if status == "yellow" {
            warn!("Elasticsearch cluster seems a bit wonky, might cause errors!");
        }

        info!(
            "Connection has been tested, received {} from /_cluster/health from cluster {}!",
            code, cluster_name
        );

        // Create the indexes
        self.create_indexes().await?;

        Ok(())
    }

    pub async fn create_indexes(&self) -> Result<(), &'static str> {
        info!("creating indexes if they don't exist...");

        let config = Config::get();
        for index in &config.elastic.indexes {
            debug!("checking if index {} exists...", index);

            let res = self
                .client
                .send(
                    Method::Head,
                    index,
                    HeaderMap::new(),
                    Option::<&Value>::None,
                    Option::<&str>::None,
                    None,
                )
                .await
                .expect("Unable to make request to Elasticsearch.");

            let status = res.status_code();
            debug!("Received status code {} on `HEAD /{}`", status, index);

            if status.is_success() {
                debug!("   => Index {} already exists! Skipping...", index);
                continue;
            }

            if status.is_server_error() {
                error!(
                    "  => Received a server error, skipping index creation (index={})",
                    index
                );
                continue;
            }

            if status.is_client_error() && status.as_u16() != 404 {
                error!(
                    " => Received client error, skipping index creation (index={})",
                    index
                );
                continue;
            }

            debug!("  => Index {} doesn't exist, now creating...", index);
            let res2 = self
                .client
                .send(
                    Method::Put,
                    index,
                    HeaderMap::new(),
                    Option::<&Value>::None,
                    Some(b"{}".as_ref()),
                    None,
                )
                .await
                .expect("Could not create index");

            let res2_status = res2.status_code();
            debug!("    => Received status {} on `PUT /{}`", res2_status, index);

            if res2_status.is_success() {
                debug!(
                    "    => Index {} now exists, you can query it using /search/:index!",
                    index
                );
            } else {
                let body = res2
                    .json::<Value>()
                    .await
                    .expect("Unable to deserialise JSON payload");

                error!("    => Unable to create index {}: {}", index, body);
            }
        }

        Ok(())
    }

    pub async fn index_exists(&self, index: &str) -> bool {
        debug!("checking if index '{}' exists...", index);

        let res = self
            .client
            .send(
                Method::Head,
                index,
                HeaderMap::new(),
                Option::<&Value>::None,
                Option::<&str>::None,
                None,
            )
            .await
            .expect("Unable to make request to Elasticsearch.");

        let status_code = res.status_code();
        debug!("Received status {} on HEAD /{}", status_code, index);

        if status_code.is_success() {
            return true;
        }

        if status_code.is_server_error() {
            return false;
        }

        if status_code.is_client_error() && status_code.as_u16() != 404 {
            return false;
        }

        // will be here if 404'd :>
        false
    }

    pub async fn index_doc_stats(&self, index: &str) -> (i64, i64) {
        debug!("get doc statistics for index '{}'!", index);

        let res = self
            .client
            .send(
                Method::Get,
                format!("/{}/_stats/docs", index).as_str(),
                HeaderMap::new(),
                Option::<&Value>::None,
                Option::<&str>::None,
                None,
            )
            .await
            .expect("Unable to make request to Elasticsearch.");

        let status_code = res.status_code();
        debug!(
            "Received status {} on GET /{}/_stats/docs",
            status_code, index
        );

        if status_code.is_server_error() {
            return (0, 0);
        }

        if status_code.is_client_error() {
            return (0, 0);
        }

        let body = res
            .json::<Value>()
            .await
            .expect("Unable to deserialise result");

        let all = body["_all"]
            .as_object()
            .expect("Unable to retrieve stats. :(");

        let total = all["total"]
            .as_object()
            .expect("Unable to retrieve stats. :(");

        let docs = total["docs"]
            .as_object()
            .expect("Unable to retrieve stats. :(");

        (
            docs["count"].as_i64().unwrap_or(0),
            docs["deleted"].as_i64().unwrap_or(0),
        )
    }

    pub async fn index_search(&self, index: &str, match_type: MatchType, data: Value) -> Value {
        debug!("now searching on index '{}'", index);
        trace!("data = {}", data);

        let payload = match match_type {
            MatchType::Fuzzy => json!({
                "query": {
                    "fuzzy": data
                }
            }),
            MatchType::MatchAll => json!({
                "query": {
                    "match_all": data
                }
            }),
        };

        let payload_string = format!("{}", payload);
        debug!("payload as string = {:?}", payload_string);

        let res = self
            .client
            .search(SearchParts::Index(&[index]))
            .body(payload)
            .human(true)
            .pretty(true)
            .send()
            .await
            .expect("Unable to create a request to Elasticsearch");

        let status_code = res.status_code();
        debug!("Received status {} on POST /{}/_search", status_code, index);

        let body = res
            .json::<Value>()
            .await
            .expect("Unable to deserialise body.");

        if status_code.is_server_error() {
            error!(
                "Received server error when searching in index '{}':\n{}",
                index, body
            );

            return json!({
                "success": false,
                "errors": [
                    {
                        "code": "INTERNAL_SERVER_ERROR",
                        "message": format!("Received >500 when requesting to `ES_CLUSTER/{}/_search`", index)
                    }
                ]
            });
        }

        if status_code.is_client_error() {
            error!(
                "Received client error when searching in index '{}':\n{}",
                index, body
            );

            return json!({
                "success": false,
                "errors": [
                    {
                        "code": "CLIENT_ERROR",
                        "message": format!("Received >= 400 <= 500 when requesting to `ES_CLUSTER/{}/_search", index)
                    }
                ]
            });
        }

        let took = body["took"].as_i64().unwrap_or(-1);
        let hits = body["hits"]
            .as_object()
            .expect("Unable to retrieve `hits` object.");

        let default_return_value: Vec<Value> = vec![];
        return json!({
            "success": true,
            "data": {
                "took": took,
                "data": hits["hits"].as_array().unwrap_or(&default_return_value)
            }
        });
    }
}
