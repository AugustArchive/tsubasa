# 🐇 tsubasa: Microservice to define a schema and execute it in a fast environment.
# Copyright 2022 Noel <cutie@floofy.dev>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

VERSION    := $(shell cat ./version.json | jq .version | tr -d '"')
COMMIT_SHA := $(shell git rev-parse --short=8 HEAD)
BUILD_DATE := $(shell go run ./cmd/build-date/main.go)

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

ifeq ($(GOOS), linux)
	TARGET_OS ?= linux
else ifeq ($(GOOS),darwin)
	TARGET_OS ?= darwin
else ifeq ($(GOOS),windows)
	TARGET_OS ?= windows
else
	$(error System $(GOOS) is not supported at this time)
endif

EXTENSION :=
ifeq ($(TARGET_OS),windows)
	EXTENSION := .exe
endif

# Usage: `make deps`
deps:
	@echo Updating dependency tree...
	go mod tidy
	go mod download
	@echo Updated dependency tree successfully.

# Usage: `make build`
build:
	@echo Now building Tsubasa for platform $(GOOS)/$(GOARCH)!
	go build -ldflags "-s -w -X floofy.dev/tsubasa/internal.Version=${VERSION} -X floofy.dev/tsubasa/internal.CommitSHA=${COMMIT_SHA} -X \"floofy.dev/tsubasa/internal.BuildDate=${BUILD_DATE}\"" -o ./bin/tsubasa$(EXTENSION)
	@echo Successfully built the binary. Use './bin/tsubasa$(EXTENSION)' to run!

# Usage: `make clean`
clean:
	@echo Now cleaning project..
	rm -rf bin/ .profile/
	go clean
	@echo Done!

# Usage: `make fmt`
fmt:
	@echo Formatting project...
	go fmt
