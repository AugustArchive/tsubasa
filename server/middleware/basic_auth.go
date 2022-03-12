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

package middleware

import (
	"crypto/subtle"
	"floofy.dev/tsubasa/internal"
	"floofy.dev/tsubasa/internal/result"
	"floofy.dev/tsubasa/util"
	"net/http"
)

func BasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Check if Basic authentication is enabled by config
		if internal.GlobalContainer.Config.Username != nil && internal.GlobalContainer.Config.Password != nil {
			u := internal.GlobalContainer.Config.Username
			p := internal.GlobalContainer.Config.Password

			user, pass, ok := req.BasicAuth()
			if !ok {
				w.Header().Add("WWW-Authenticate", `Basic realm="Noel/Tsubasa"`)
				res := result.Err(http.StatusUnauthorized, "UNABLE_TO_OBTAIN", "Server has enabled basic authentication and I couldn't grab the credentials. :(")

				util.WriteJson(w, http.StatusUnauthorized, res)
				return
			}

			if user != *u {
				w.Header().Add("WWW-Authenticate", `Basic realm="Noel/Tsubasa"`)
				res := result.Err(http.StatusUnauthorized, "INVALID_USERNAME", "Invalid username.")

				util.WriteJson(w, http.StatusUnauthorized, res)
				return
			}

			if subtle.ConstantTimeCompare([]byte(*p), []byte(pass)) != 1 {
				w.Header().Add("WWW-Authenticate", `Basic realm="Noel/Tsubasa"`)
				res := result.Err(http.StatusUnauthorized, "INVALID_PASSWORD", "Invalid password.")

				util.WriteJson(w, http.StatusUnauthorized, res)
				return
			}

			next.ServeHTTP(w, req)
		} else {
			next.ServeHTTP(w, req)
		}
	})
}
