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

package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/vanus-labs/vanus-operator/api/models"
	"github.com/vanus-labs/vanus-operator/api/restapi/operations/healthz"
)

func RegistHealthzHandler(a *Api) {
	a.HealthzHealthzHandler = healthz.HealthzHandlerFunc(a.healthzHandler)
}

func (a *Api) healthzHandler(params healthz.HealthzParams) middleware.Responder {
	retcode := int32(200)
	msg := "ok"
	return healthz.NewHealthzOK().WithPayload(&healthz.HealthzOKBody{
		Code: &retcode,
		Data: &models.HealthInfo{
			Status: "OK",
		},
		Message: &msg,
	})
}
