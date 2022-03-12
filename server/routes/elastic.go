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
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func NewElasticRouter() chi.Router {
	r := chi.NewRouter()
	elastic := internal.GlobalContainer.Elastic

	r.Get("/{index}", func(w http.ResponseWriter, req *http.Request) {
		exists := elastic.IndexExists(chi.URLParam(req, "id"))
		util.WriteJson(w, 200, result.Ok(map[string]interface{}{
			"exists": exists,
		}))
	})

	r.Post("/{index}/search", func(w http.ResponseWriter, req *http.Request) {
		status, body, err := util.GetJsonBody(req)
		if err != nil {
			util.WriteJson(w, status, result.Err(status, "INVALID_JSON_BODY", err.Error()))
			return
		}

		index := chi.URLParam(req, "index")
		matchType, ok := body["match_type"].(string)
		if !ok {
			util.WriteJson(w, 406, result.Err(406, "INVALID_DATA_TYPE", fmt.Sprintf("Invalid data type on {match_type=>%v} (expected string)", matchType)))
			return
		}

		data, ok := body["data"].(map[string]interface{})
		if !ok {
			util.WriteJson(w, 406, result.Err(406, "INVALID_DATA_TYPE", fmt.Sprintf("Invalid data type on {data=>%v} (expected JSON object)", matchType)))
			return
		}

		res := elastic.SearchInIndex(index, matchType, data)
		util.WriteJson(w, res.StatusCode, res)
	})

	return r
}
