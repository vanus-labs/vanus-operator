// Copyright 2017 EasyStack, Inc.
package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/linkall-labs/vanus-operator/api/models"
	"github.com/linkall-labs/vanus-operator/api/restapi/operations/healthz"
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
