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

# This is the dockerfile for development if you need it. :shrug:
FROM golang:1.17-alpine AS builder

# Install the needed dependencies
RUN apk update && apk add --no-cache ca-certificates git make jq

# Change the working directory to /build/ume
WORKDIR /build/tsubasa
COPY . .

# Build the source code
RUN make deps
RUN make build

# Now, this is the final stage! :D
FROM alpine:3.15

# install needed dependencies
RUN apk update && apk add --no-cache bash musl-dev libc-dev gcompat

# Change the directory to `/app/noel/tsubasa`
WORKDIR /app/noel/tsubasa

# Bring in our Docker scripts to the `scripts/` directory
COPY docker /app/noel/tsubasa/scripts
COPY --from=builder /build/tsubasa/bin/tsubasa .

RUN chmod +x /app/noel/tsubasa/scripts/docker-entrypoint.sh \
  /app/noel/tsubasa/scripts/runner.sh

RUN ln -s /app/noel/tsubasa/tsubasa /usr/bin/tsubasa

USER 1001

ENTRYPOINT ["/app/noel/tsubasa/scripts/docker-entrypoint.sh"]
CMD ["/app/noel/tsubasa/scripts/runner.sh"]
