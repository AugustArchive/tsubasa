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

package internal

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
	"os"
)

// GlobalContainer represents the global container that is initialised by
// calling NewContainer.
var GlobalContainer *Container = nil

// Container is the container to hold any object any other package requires.
type Container struct {
	// Represents the service for handling Elastic-related objects.
	Elastic *ElasticService

	// Represents the Sentry client if enabled.
	Sentry *sentry.Client

	// Represents the configuration that was loaded
	Config *Config
}

// NewContainer creates a new Container object and initializes the GlobalContainer
// variable. If the GlobalContainer is already defined, it will panic.
func NewContainer(config *Config) *Container {
	if GlobalContainer != nil {
		panic("Unable to create new container since one is already created.")
	}

	elastic, err := NewElasticService(config)
	if err != nil {
		logrus.Fatalf("Unable to create a connection to Elasticsearch: %v", err)
	}

	var sc *sentry.Client
	if config.SentryDSN != nil {
		logrus.Info("Sentry logging is enabled, now installing...")
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "localhost"
		}

		client, err := sentry.NewClient(sentry.ClientOptions{
			Dsn:              *config.SentryDSN,
			AttachStacktrace: true,
			SampleRate:       1.0,
			ServerName:       fmt.Sprintf("noel.tsubasa v%s @ %s", Version, hostname),
		})

		if err != nil {
			logrus.Fatalf("Unable to create Sentry client: %s", err)
		}

		sc = client
	}

	GlobalContainer = &Container{
		Elastic: elastic,
		Sentry:  sc,
		Config:  config,
	}

	return GlobalContainer
}
