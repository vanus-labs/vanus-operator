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
testmode=${COVERMODE:-atomic}
testdir=$(mktemp -d /tmp/test.XXXXXXXXXX)
profile="${testdir}/test.out"
hash godir 2>/dev/null || go get github.com/Masterminds/godir
generate_test_data() {
  # echo "mode: $testmode" >"$profile"
  echo "Start run golang unittest check"
  for d in $(godir pkg) ; do
    (
      # go test "$d" >>"$profile"
      go test "$d"
    )
  done
}
generate_test_data
