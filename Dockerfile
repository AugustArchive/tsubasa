FROM rustlang/rust:nightly-alpine3.15 AS builder

# why is rust like this
# source: https://github.com/Benricheson101/anti-phishing-bot/blob/single_server/Dockerfile
RUN apk update && apk add --no-cache build-base openssl-dev gcompat libc6-compat
WORKDIR /build/tsubasa

# This basically builds all the dependencies that Tsubasa requires
COPY Cargo.toml .
RUN echo "fn main() {}" >> dummy.rs
RUN sed -i 's#src/main.rs#dummy.rs#' Cargo.toml
ENV RUSTFLAGS=-Ctarget-feature=-crt-static
RUN CARGO_INCREMENTAL=1 cargo build --release
RUN rm dummy.rs && sed -i 's#dummy.rs#src/main.rs#' Cargo.toml

# Now we build the actual server
COPY . .
RUN cargo build --release

# This is the main thing that will be ran, multi-stage builds ftw!
FROM alpine:3.15

# ARG VERSION
# ARG COMMIT_HASH
# ARG BUILD_DATE

# add external metadata!
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

# Install needed dependencies
WORKDIR /app/noel/tsubasa
RUN apk update && apk add --no-cache build-base openssl bash
COPY docker /app/noel/tsubasa/scripts
COPY --from=builder /build/tsubasa/target/release/tsubasa .

RUN chmod +x /app/noel/tsubasa/scripts/docker-entrypoint.sh \
  /app/noel/tsubasa/scripts/runner.sh

USER 1001

ENTRYPOINT ["/app/noel/tsubasa/scripts/docker-entrypoint.sh"]
CMD ["/app/noel/tsubasa/scripts/runner.sh"]
