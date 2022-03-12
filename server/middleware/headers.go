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
	"floofy.dev/tsubasa/internal"
	"fmt"
	"net/http"
)

func Headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.Header.Set("X-Powered-By", fmt.Sprintf("Noel/Tsubasa v%s", internal.Version))
		req.Header.Set("Cache-Control", "public, max-age=7776000")

		req.Header.Set("X-Frame-Options", "deny")
		req.Header.Set("X-Content-Type-Options", "nosniff")
		req.Header.Set("X-XSS-Protection", "1; mode=block")

		next.ServeHTTP(w, req)
	})
}
