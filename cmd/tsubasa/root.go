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

package tsubasa

import (
	"floofy.dev/tsubasa/internal"
	"floofy.dev/tsubasa/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command
var verbose *bool

func init() {
	// Initialise command
	rootCmd = &cobra.Command{
		Use:   "tsubasa [COMMAND] [ARGS...]",
		Short: "Manages the `tsubasa` service.",
		Long: `Tsubasa is a simple microservice to abstract Elasticsearch so you can focus on bringing
a good application without bringing the Elastic SDK into your application.
`,
		RunE: runServer,
	}

	// Set persisted flags
	rootCmd.PersistentFlags().StringP("config", "c", "", "The configuration file to bootstrap the server.")
	verbose = rootCmd.PersistentFlags().BoolP("verbose", "v", false, "If verbose mode should be enabled (overrides `config.debug`)")
	rootCmd.AddCommand(newGenerateCommand())
}

func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		return 1
	}

	return 0
}

func runServer(_ *cobra.Command, _ []string) error {
	configPath := rootCmd.Flag("config").Value.String()

	var path *string = nil
	if configPath != "" {
		path = &configPath
	}

	if verbose != nil && *verbose == true {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.SetFormatter(internal.NewFormatter())
	logrus.SetReportCaller(true)

	var config *internal.Config
	if path == nil {
		c, err := internal.FindAndNewConfig()
		if err != nil {
			panic(err)
		}

		config = c
	} else {
		c, err := internal.NewConfig(*path)
		if err != nil {
			panic(err)
		}

		config = c
	}

	if config.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	return server.Start(config)
}
