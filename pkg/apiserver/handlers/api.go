// Copyright 2017 EasyStack, Inc.
package handlers

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/linkall-labs/vanus-operator/api/restapi"
	"github.com/linkall-labs/vanus-operator/api/restapi/operations"
	"github.com/linkall-labs/vanus-operator/pkg/apiserver/controller"

	"github.com/go-kit/kit/log"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"

	"net/http"
)

type Api struct {
	*operations.VanusAPI
	basepath string
	ctx      context.Context
	log      log.Logger
	ctrl     *controller.Controller
}

func NewApi(log log.Logger, ctrl *controller.Controller, basepath string) (*Api, error) {
	// Load embedded swagger file.
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded swagger file: %v", err.Error())
	}

	// Create new service API.
	openAPI := operations.NewVanusAPI(swaggerSpec)
	swaggerSpec.Spec().BasePath = basepath
	// Skip the  redoc middleware, only serving the OpenAPI specification and
	// the API itself via RoutesHandler. See:
	// https://github.com/go-swagger/go-swagger/issues/1779
	// openAPI.Middleware = func(b middleware.Builder) http.Handler {
	// 	return middleware.Spec(swaggerSpec.Spec().BasePath, swaggerSpec.Raw(), openAPI.Context().RoutesHandler(b))
	// }

	openAPI.Middleware = func(b middleware.Builder) http.Handler {
		return middleware.Spec(swaggerSpec.Spec().BasePath, swaggerSpec.Raw(),
			middleware.SwaggerUI(middleware.SwaggerUIOpts{
				BasePath: basepath,
				Path:     "swaggerui",
				SpecURL:  filepath.Join(basepath, "swagger.json"),
			}, openAPI.Context().RoutesHandler(b)))
	}

	api := &Api{
		log:      log,
		VanusAPI: openAPI,
		ctrl:     ctrl,
		basepath: swaggerSpec.Spec().BasePath,
		ctx:      context.Background(),
	}

	RegistClusterHandler(api)
	RegistConnectorHandler(api)
	RegistHealthzHandler(api)
	return api, nil
}

func (a *Api) BasePath() string {
	return a.basepath
}

func (a *Api) Handler() http.Handler {
	return a.VanusAPI.Serve(nil)
}
