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

FROM alpine:3.15

# Set the working directory + install dependencies
RUN apk update && apk add --no-cache build-base openssl bash
WORKDIR /app/noel/tsubasa

# Apply extra metadata on this image
LABEL MAINTAINER="Noel <cutie@floofy.dev>"
LABEL gay.floof.tsubasa.version=${VERSION}
LABEL gay.floof.tsubasa.commitSha=${COMMIT_HASH}
LABEL gay.floof.tsubasa.buildDate=${BUILD_DATE}
LABEL org.opencontainers.image.title="tsubasa"
LABEL org.opencontainers.image.description="Tiny microservice to define a schema and then be executed by any search engine you wish to use, like Elasticsearch, Meilisearch, or OpenSearch!"
LABEL org.opencontainers.image.source=https://github.com/auguwu/tsubasa
# LABEL org.opencontainers.image.version=${VERSION}
# LABEL org.opencontainers.image.created=${BUILD_DATE}
# LABEL org.opencontainers.image.revision=${COMMIT_HASH}
LABEL org.opencontainers.image.licenses="Apache-2.0"

# Copy our Docker scripts to this directory
COPY docker /app/noel/tsubasa/scripts
COPY ./target/release/tsubasa .

# Make these scripts executable
RUN chmod +x /app/noel/tsubasa/scripts/docker-entrypoint.sh \
  /app/noel/tsubasa/scripts/runner.sh

# Use a non-root context.
USER 1001

# Set our entrypoints
ENTRYPOINT ["/app/noel/tsubasa/scripts/docker-entrypoint.sh"]
CMD ["/app/noel/tsubasa/scripts/runner.sh"]
