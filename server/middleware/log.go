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
	"floofy.dev/tsubasa/util"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		s := time.Now()
		ww := middleware.NewWrapResponseWriter(w, req.ProtoMajor)
		next.ServeHTTP(ww, req)

		code := util.GetStatusCode(ww.Status())
		logrus.Infof("[%s] %s %s (%s) => %d %s (%d bytes [%s])",
			req.RemoteAddr,
			req.Method,
			req.URL.Path,
			req.Proto,
			ww.Status(),
			code,
			ww.BytesWritten(),
			time.Since(s).String())
	})
}
