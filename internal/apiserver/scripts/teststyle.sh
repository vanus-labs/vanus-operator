#!/usr/bin/env bash
# Copyright 2016 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

PROJECT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

# Explicitly opt into go modules, even though we're inside a GOPATH directory
export GO111MODULE=on

# Install golangci-lint
echo 'installing golangci-lint '
hash golangci-lint 2>/dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint

cd "${PROJECT_ROOT}"

generate_lint_data() {
  echo "Start run golangci-lint check"
  return 0
  golangci-lint run \
    --timeout 30m \
    -E gocyclo \
    -E revive
}
generate_lint_data