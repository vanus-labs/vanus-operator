// Copyright 2017 EasyStack, Inc.
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
	"github.com/linkall-labs/vanus-operator/api/models"
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
