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
cyclomode=${COVERMODE:-atomic}
cyclodir=$(mktemp -d /tmp/cyclo.XXXXXXXXXX)
profile="${cyclodir}/cyclo.out"
hash gocyclo 2>/dev/null || go get github.com/fzipp/gocyclo/cmd/gocyclo
hash godir 2>/dev/null || go get github.com/Masterminds/godir
generate_cyclo_data() {
  # echo "mode: $cyclomode" >"$profile"
  echo "Start run golang cyclomatic complexities check"
  for d in $(godir pkg) ; do
    (
      # gocyclo -over 3 "${GOPATH}src/$d" >>"$profile"
      gocyclo -over 3 "${GOPATH}src/$d"
    )
  done
}
generate_cyclo_data
exit 0
