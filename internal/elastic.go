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
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"floofy.dev/tsubasa/internal/result"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

// MatchType represents the match type for ElasticService.SearchInIndex.
type MatchType string

var (
	FUZZY    MatchType = "fuzzy"
	MatchAll MatchType = "match_all"
	UNKNOWN  MatchType = "?"
)

func (s MatchType) String() string {
	switch s {
	case FUZZY:
		return "fuzzy"

	case MatchAll:
		return "match_all"

	default:
		return "?"
	}
}

func DetermineMatchType(s string) MatchType {
	if s == "fuzzy" || s == "Fuzzy" {
		return FUZZY
	}

	if s == "match_all" || s == "MatchAll" || s == "match" {
		return MatchAll
	}

	return UNKNOWN
}

type ElasticService struct {
	indexes []string
	client  *elasticsearch.Client
}

func NewElasticService(config *Config) (*ElasticService, error) {
	logrus.Info("Now connecting to Elasticsearch...")

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = 10

	cfg := elasticsearch.Config{
		Addresses:            config.Elastic.Nodes,
		DiscoverNodesOnStart: true,
	}

	if config.Elastic.Username != nil {
		cfg.Username = *config.Elastic.Username
	}

	if config.Elastic.Password != nil {
		cfg.Password = *config.Elastic.Password
	}

	if config.Elastic.SkipSSLVerify {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	if config.Elastic.CACertPath != nil {
		logrus.Infof("Specified TLS certificate for Elastic at path %v!", config.Elastic.CACertPath)

		var err error
		if transport.TLSClientConfig.RootCAs, err = x509.SystemCertPool(); err != nil {
			logrus.Fatal("Unable to assign root certificates:", err)
		}

		cacert, err := ioutil.ReadFile(*config.Elastic.CACertPath)
		if err != nil {
			return nil, err
		}

		transport.TLSClientConfig.RootCAs.AppendCertsFromPEM(cacert)
		transport.TLSClientConfig.ClientAuth = tls.RequireAnyClientCert
		transport.TLSClientConfig.InsecureSkipVerify = true
	}

	cfg.Transport = transport
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	// Check if we can query from the server!
	res, err := client.Info()
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var data map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	version := data["version"].(map[string]interface{})["number"].(string)
	logrus.Debugf("Server: %s | Client: %s", version, elasticsearch.Version)

	service := &ElasticService{config.Elastic.Indexes, client}
	service.createIndexes()

	return service, nil
}

func (es *ElasticService) createIndexes() {
	logrus.Info("Now creating indices if not found...")

	for _, index := range es.indexes {
		logrus.Infof("   => Checking if index %s exists...", index)
		res, err := es.client.Indices.Exists([]string{index}, es.client.Indices.Exists.WithErrorTrace())

		if err != nil {
			logrus.Errorf("  => Received error (%v) while checking if index %s exists, skipping!", err, index)
			continue
		}

		if res.StatusCode == 404 {
			logrus.Debugf("  => Index %s does not exist, now creating...", index)
			_, err := es.client.Indices.Create(index, es.client.Indices.Create.WithErrorTrace())
			if err != nil {
				logrus.Errorf("    => Unable to create index %s: %v", index, err)
				continue
			}

			logrus.Infof("    => Index %s is created!", index)
		}
	}
}

func (es *ElasticService) Available() (bool, int64) {
	logrus.Debug("Checking if Elasticsearch is available to be used")

	t := time.Now()
	res, err := es.client.Ping()

	if err != nil {
		return false, -1
	}

	if res.IsError() {
		return false, -1
	} else {
		return true, time.Since(t).Milliseconds()
	}
}

func (es *ElasticService) IndexExists(index string) bool {
	logrus.Debug("Checking if index %s exists...", index)

	res, err := es.client.Indices.Exists([]string{index}, es.client.Indices.Exists.WithErrorTrace())
	if err != nil {
		return false
	}

	if res.IsError() {
		return false
	} else {
		return true
	}
}

func (es *ElasticService) SearchInIndex(index string, matchType string, data interface{}) *result.Result {
	// Determine the match type right now
	match := DetermineMatchType(matchType)
	if match == UNKNOWN {
		return result.Err(406, "INVALID_MATCH_TYPE", fmt.Sprintf("Match type '%s' is not a valid match type.", matchType))
	}

	logrus.Debugf("Now searching data on index '%s'...", index)
	logrus.Tracef("data to search => %v", data)

	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			match.String(): data,
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		logrus.Errorf("Unable to encode query %v: %v", query, err)
		return result.Err(500, "INTERNAL_SERVER_ERROR", "Unknown service error has occurred.")
	}

	t := time.Now()
	res, err := es.client.Search(
		es.client.Search.WithIndex(index),
		es.client.Search.WithContext(context.Background()),
		es.client.Search.WithBody(&buf),
		es.client.Search.WithTrackTotalHits(true))

	if err != nil {
		logrus.Errorf("Unable to encode query %v: %v", query, err)
		return result.Err(500, "INTERNAL_SERVER_ERROR", "Unknown service error has occurred.")
	}

	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			logrus.Errorf("Unable to decode JSON payload from Elastic when received a non-acceptable status code: %s", err)
			return result.Err(500, "INTERNAL_SERVER_ERROR", "Unknown service error has occurred.")
		} else {
			logrus.Errorf("Unable to search data (%v) from index %s because: '%s'.",
				data,
				index,
				fmt.Sprintf("%s: %s",
					e["error"].(map[string]interface{})["type"],
					e["error"].(map[string]interface{})["reason"]))

			return result.Err(500, "INTERNAL_SERVER_ERROR", "Unknown service error has occurred.")
		}
	}

	var d map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&d); err != nil {
		logrus.Errorf("Unable to decode JSON payload from Elastic: %s", err)
		return result.Err(500, "INTERNAL_SERVER_ERROR", "Unknown service error has occurred.")
	}

	since := time.Since(t).Milliseconds()
	took := d["took"].(float64)
	hits := d["hits"].(map[string]interface{})
	maxScore, ok := hits["max_score"].(float64)
	if !ok {
		maxScore = float64(0)
	}

	totalHits := hits["total"].(map[string]interface{})["value"].(float64)
	rawHitsData, ok := hits["hits"].([]map[string]interface{})
	if !ok {
		rawHitsData = make([]map[string]interface{}, 0)
	}

	actualData := make([]interface{}, 0)
	if rawHitsData != nil {
		for _, hit := range rawHitsData {
			// Get the source
			source := hit["_source"]
			actualData = append(actualData, source)
		}
	}

	return result.Ok(map[string]interface{}{
		"request_ms": since,
		"took":       took,
		"max_score":  maxScore,
		"total_hits": totalHits,
		"data":       actualData,
	})
}

func (es ElasticService) SearchRaw(index string, data map[string]interface{}) *result.Result {
	logrus.Debugf("Now searching data on index '%s'...", index)
	logrus.Tracef("data to search => %v", data)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		logrus.Errorf("Unable to encode query %v: %v", data, err)
		return result.Err(500, "INTERNAL_SERVER_ERROR", "Unknown service error has occurred.")
	}

	t := time.Now()
	res, err := es.client.Search(
		es.client.Search.WithIndex(index),
		es.client.Search.WithContext(context.Background()),
		es.client.Search.WithBody(&buf),
		es.client.Search.WithTrackTotalHits(true))

	if err != nil {
		logrus.Errorf("Unable to encode query %v: %v", data, err)
		return result.Err(500, "INTERNAL_SERVER_ERROR", "Unknown service error has occurred.")
	}

	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			logrus.Errorf("Unable to decode JSON payload from Elastic when received a non-acceptable status code: %s", err)
			return result.Err(500, "INTERNAL_SERVER_ERROR", "Unknown service error has occurred.")
		} else {
			logrus.Errorf("Unable to search data (%v) from index %s because: '%s'.",
				data,
				index,
				fmt.Sprintf("%s: %s",
					e["error"].(map[string]interface{})["type"],
					e["error"].(map[string]interface{})["reason"]))

			return result.Err(500, "INTERNAL_SERVER_ERROR", "Unknown service error has occurred.")
		}
	}

	var d map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&d); err != nil {
		logrus.Errorf("Unable to decode JSON payload from Elastic: %s", err)
		return result.Err(500, "INTERNAL_SERVER_ERROR", "Unknown service error has occurred.")
	}

	since := time.Since(t).Milliseconds()
	took := d["took"].(float64)
	hits := d["hits"].(map[string]interface{})
	maxScore, ok := hits["max_score"].(float64)
	if !ok {
		maxScore = float64(0)
	}

	totalHits := hits["total"].(map[string]interface{})["value"].(float64)
	rawHitsData, ok := hits["hits"].([]map[string]interface{})
	if !ok {
		rawHitsData = make([]map[string]interface{}, 0)
	}

	actualData := make([]interface{}, 0)
	if rawHitsData != nil {
		for _, hit := range rawHitsData {
			// Get the source
			source := hit["_source"]
			actualData = append(actualData, source)
		}
	}

	return result.Ok(map[string]interface{}{
		"request_ms": since,
		"took":       took,
		"max_score":  maxScore,
		"total_hits": totalHits,
		"data":       actualData,
	})
}
