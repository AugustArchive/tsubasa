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
	"github.com/pelletier/go-toml/v2"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

// Config represents the configuration for Tsubasa. The configuration file
// should be located under:
//        - **TSUBASA_CONFIG_PATH** environment variable;
//        - $ROOT/config.toml
//        - `/app/noel/tsubasa/config.toml` or `C:\\User\\<username>\\AppData\\Local\\tsubasa\\config.toml`
type Config struct {
	// If Sentry logging should be enabled on the server.
	SentryDSN *string `toml:"sentry_dsn"`

	// The username to use to enable Basic authentication on the Tsubasa server.
	// This requires the `password` field to be defined!
	Username *string `toml:"username"`

	// The password to use to enable Basic authentication on the Tsubasa server.
	// This requires the `username` field to be defined.
	Password *string `toml:"password"`

	// The configuration to use to configure Elasticsearch.
	Elastic ElasticConfig `toml:"elastic"`

	// If debug logging should be enabled.
	Debug bool `toml:"debug"`

	// The host to use that Tsubasa should listen to. By default, it will
	// listen at 0.0.0.0
	Host *string `toml:"host"`

	// The HTTP port that Tsubasa should be listening to. By default,
	// it allocates the port: 23145
	Port *int `toml:"port"`
}

type ElasticConfig struct {
	// The password to use if Basic authentication is enabled on the server.
	Password *string `toml:"password,omitempty"`

	// The username to use if Basic authentication is enabled on the server.
	Username *string `toml:"username,omitempty"`

	// The list of indexes Tsubasa should keep track of. If the index doesn't exist,
	// then Tsubasa will configure it.
	Indexes []string `toml:"indexes"`

	// The list of nodes to use when connecting to Elasticsearch.
	Nodes []string `toml:"nodes"`

	// CACertPath is the path to a .pem file to use TLS connections within
	// Elasticsearch.
	CACertPath *string `toml:"ca_path,omitempty"`

	// SkipSSLVerify skips the SSL certificates.
	SkipSSLVerify bool `toml:"skip_ssl_verify"`
}

// NewConfig initialized the configuration for Tsubasa.
func NewConfig(path string) (*Config, error) {
	logrus.Infof("Loading configuration from path '%s'...", path)

	// Now, let's read it
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.Fatalf("Unable to read from path '%s': %v", path, err)
	}

	var config Config
	if err := toml.Unmarshal(contents, &config); err != nil {
		logrus.Fatalf("Unable to unmarshal config from path '%s': %v", path, err)
	}

	logrus.Info("Loaded configuration successfully. :)")
	return &config, nil
}

// FindAndNewConfig finds the configuration under the following paths and loads it.
//        - **TSUBASA_CONFIG_PATH** environment variable;
//        - $ROOT/config.toml
//        - `/app/noel/tsubasa/config.toml` or `C:\\User\\<username>\\AppData\\Local\\tsubasa\\config.toml`
func FindAndNewConfig() (*Config, error) {
	logrus.Debug("Finding configuration path...")

	// Let's figure out the path!
	path := ""

	if p, ok := os.LookupEnv("TSUBASA_CONFIG_PATH"); ok {
		logrus.Debugf("   => Found config path under `TSUBASA_CONFIG_PATH`: %s", p)
		path = p
	}

	// Check if the `path` variable wasn't filled in
	if path == "" {
		logrus.Debug("   => Could not find it under `TSUBASA_CONFIG_PATH`, checking $ROOT/config.toml")
		_, err := os.Stat("./config.toml")

		if !os.IsNotExist(err) {
			logrus.Debugf("   => Found it under the current directory!")
			path = "./config.toml"
		} else {
			logrus.Debugf("   => Could not find it under the current directory, assuming it's under /app/noel/tsubasa/config.toml")
			path = "/app/noel/tsubasa/config.toml"
		}
	}

	return NewConfig(path)
}
