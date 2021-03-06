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

package routes

import (
	"floofy.dev/tsubasa/internal"
	"floofy.dev/tsubasa/internal/result"
	"floofy.dev/tsubasa/util"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func NewMainRouter() chi.Router {
	r := chi.NewRouter()
	elastic := internal.GlobalContainer.Elastic

	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		util.WriteJson(w, 200, result.Ok(map[string]any{
			"hello": "world",
		}))
	})

	r.Get("/info", func(w http.ResponseWriter, req *http.Request) {
		healthy, ping := elastic.Available()

		util.WriteJson(w, 200, result.Ok(map[string]any{
			"version":    internal.Version,
			"commit_sha": internal.CommitSHA,
			"build_date": internal.BuildDate,
			"elastic": map[string]any{
				"healthy":        healthy,
				"ping":           ping,
				"server_version": elastic.ServerVersion,
				"client_version": elasticsearch.Version,
			},
		}))
	})

	return r
}
