#!/bin/bash

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

set -o errexit
set -o nounset
set -o pipefail

. /app/noel/tsubasa/scripts/liblog.sh

if ! [[ "${TSUBASA_ENABLE_WELCOME_PROMPT:-yes}" =~ ^(no|false)$ ]]; then
    info ""
    info "   Welcome to the ${BOLD}tsubasa${RESET} container image."
    info "   Tiny, and simple Elasticsearch microservice to abstract searching objects!"
    info ""
    info "   Subscribe to the project for more updates: https://github.com/auguwu/tsubasa"
    info "   Any issues occur? Report it!               https://github.com/auguwu/tsubasa/issues"
    info ""
fi

exec "$@"
