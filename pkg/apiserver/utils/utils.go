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

package utils

import (
	"bytes"
	"net/http"
	"strings"

	"encoding/json"
	"math/rand"
	"sync"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/vanus-labs/vanus-operator/api/models"
	"k8s.io/klog/v2"
)

var (
	letters = []rune("0123456789abcdefghijklmnopqrstuvwxyz")
)

func RandStr(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

var (
	DefaultGetTimeout = time.Second * 10

	bufpool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

func GetBuf() *bytes.Buffer {
	buf := bufpool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

func PutBuf(b *bytes.Buffer) {
	bufpool.Put(b)
}

type StrIter map[string]StrIter

func NewStrIter(s string) StrIter {
	var (
		ret = map[string]StrIter{}
	)
	slist := strings.Split(s, ".")
	for _, v := range slist {
		ret[v] = map[string]StrIter{}
		ret = ret[v]
	}
	return ret
}

func Response(code int32, err error) middleware.Responder {
	var (
		message string
	)
	if code == 0 {
		code = 400
	}
	if err != nil {
		klog.Errorf("HANDLE Failed: %v", err)
		message = err.Error()
	} else {
		message = "failed"
	}
	resp := models.APIResponse{
		Code:    code,
		Message: message,
	}
	respData, _ := json.Marshal(resp)
	return middleware.ResponderFunc(func(resp http.ResponseWriter, _ runtime.Producer) {
		resp.WriteHeader(http.StatusOK)
		resp.Write(respData) //nolint
	})
}
