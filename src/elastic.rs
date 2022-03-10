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

use crate::config::Config;
use elasticsearch::{
    auth::Credentials,
    cluster::ClusterHealthParts,
    http::{
        transport::{SingleNodeConnectionPool, TransportBuilder},
        Url,
    },
    Elasticsearch as ES,
};
use once_cell::sync::OnceCell;
use serde_json::Value;

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

    pub fn new() {
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

        if auth.is_some() {
            transport_builder = transport_builder.auth(auth.unwrap());
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

        if status != "green" {
            return Err(
                "Cluster is not healthy, please restart Tsubasa once the ES cluster is available!",
            );
        }

        info!(
            "Connection has been tested, received {} from /_cluster/health from cluster {}!",
            code, cluster_name
        );

        Ok(())
    }
}
