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
	"context"
	"floofy.dev/tsubasa/internal"
	"floofy.dev/tsubasa/internal/result"
	"floofy.dev/tsubasa/util"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func ErrorHandling(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// If sentry isn't enabled, let's just log the panic
		if internal.GlobalContainer.Sentry == nil {
			defer func() {
				if err := recover(); err != nil {
					if err == http.ErrAbortHandler {
						panic(err)
					}

					logrus.Errorf("Received panic on rotue '%s %s':", req.Method, req.URL.Path)
					middleware.PrintPrettyStack(err)

					res := result.Err(500, "INTERNAL_SERVER_ERROR", "Unknown service error has occurred.")
					util.WriteJson(w, 500, res)
				}
			}()

			next.ServeHTTP(w, req)
		} else {
			ctx := req.Context()
			hub := sentry.GetHubFromContext(ctx)
			if hub == nil {
				hub = sentry.CurrentHub().Clone()
				ctx = sentry.SetHubOnContext(ctx, hub)
			}

			span := sentry.StartSpan(ctx, "noel.tsubasa.service.request",
				sentry.TransactionName(fmt.Sprintf("request -> %s %s", req.Method, req.URL.Path)),
				sentry.ContinueFromRequest(req))

			defer span.Finish()

			req = req.WithContext(span.Context())
			hub.Scope().SetRequest(req)
			defer func() {
				if err := recover(); err != nil {
					if err == http.ErrAbortHandler {
						panic(err)
					}

					logrus.Errorf("Received panic on rotue '%s %s':", req.Method, req.URL.Path)
					middleware.PrintPrettyStack(err)

					eventId := hub.RecoverWithContext(context.WithValue(req.Context(), sentry.RequestContextKey, req), err)
					if eventId != nil {
						hub.Flush(1 * time.Second)
					}

					res := result.Err(500, "INTERNAL_SERVER_ERROR", "Unknown service error has occurred.")
					util.WriteJson(w, 500, res)
				}
			}()

			next.ServeHTTP(w, req)
		}
	})
}
