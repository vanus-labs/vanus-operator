// Copyright 2017 EasyStack, Inc.
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
