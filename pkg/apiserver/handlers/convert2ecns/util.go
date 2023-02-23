// Copyright 2023 Linkall Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package convert2ecns

import (
	"io"
	"time"

	"k8s.io/client-go/util/jsonpath"
)

var (
	parse *jsonpath.JSONPath
)

func init() {
	parse = jsonpath.New("")
}

func PtrS(s string) *string {
	return &s
}

func PtrInt32(s int32) *int32 {
	return &s
}

func PtrBool(s bool) *bool {
	return &s
}

func Format(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func Parse(data interface{}, w io.Writer) error {
	return parse.Execute(w, data)
}
