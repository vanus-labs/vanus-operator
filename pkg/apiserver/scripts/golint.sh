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
set -uo pipefail
lintmode=${COVERMODE:-atomic}
lintdir=$(mktemp -d /tmp/lint.XXXXXXXXXX)
profile="${lintdir}/lint.out"
hash golint 2>/dev/null || go get golang.org/x/lint/golint
hash godir 2>/dev/null || go get github.com/Masterminds/godir
generate_lint_data() {
  # echo "mode: $lintmode" >"$profile"
  echo "Start run golint check"
  for d in $(godir pkg) ; do
    (
      golint "${GOPATH}src/$d"
    )
  done
  # grep -h -v "^mode:" "$profile" >/dev/null 2>&1
  # if [ $? -eq 0 ]; then exit 1; fi
}
generate_lint_data
