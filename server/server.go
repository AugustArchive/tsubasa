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

package server

import (
	"context"
	"floofy.dev/tsubasa/internal"
	"floofy.dev/tsubasa/internal/result"
	"floofy.dev/tsubasa/server/middleware"
	"floofy.dev/tsubasa/server/routes"
	"floofy.dev/tsubasa/util"
	"fmt"
	"github.com/go-chi/chi/v5"
	chim "github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Start(config *internal.Config) error {
	container := internal.NewContainer(config)

	if internal.Root() {
		logrus.Warn("Make sure you are not running Tsubasa under administration privileges.")
	}

	router := chi.NewRouter()

	// Define global middleware before the chain starts
	router.NotFound(func(w http.ResponseWriter, req *http.Request) {
		res := result.Err(404, "ROUTE_NOT_FOUND", fmt.Sprintf("Unknown route \"%s %s\"! Are you in the right path? :blbctscrd:", req.Method, req.URL.Path))
		util.WriteJson(w, 404, res)
	})

	router.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		res := result.Err(405, "INVALID_METHOD", fmt.Sprintf(":blbctscrd: Using method %s on route %s? Sorry, wrong one! :<", req.Method, req.URL.Path))
		util.WriteJson(w, 405, res)
	})

	// Define middleware and routing here
	router.Use(chim.RealIP)
	router.Use(chim.GetHead)
	router.Use(middleware.Logging)
	router.Use(middleware.Headers)
	router.Use(middleware.BasicAuth)
	router.Use(middleware.ErrorHandling)
	router.Mount("/", routes.NewMainRouter())
	router.Mount("/health", routes.NewHealthRouter())
	router.Mount("/elastic", routes.NewElasticRouter())

	port := 23145
	if container.Config.Port != nil {
		port = *container.Config.Port
	}

	host := "0.0.0.0"
	if container.Config.Host != nil {
		host = *container.Config.Host
	}

	address := fmt.Sprintf("%s:%d", host, port)
	server := &http.Server{
		Addr:         address,
		Handler:      router,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logrus.Infof("Tsubasa is now listening under address => %s", address)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logrus.Errorf("Unable to run HTTP server: %s", err)
		}
	}()

	<-sigint

	logrus.Warn("Shutting down HTTP server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	go func() {
		<-shutdownCtx.Done()
		if shutdownCtx.Err() == context.DeadlineExceeded {
			logrus.Warn("")
		}
	}()

	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	} else {
		return nil
	}
}
