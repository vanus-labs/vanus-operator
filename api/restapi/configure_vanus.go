// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/vanus-labs/vanus-operator/api/restapi/operations"
	"github.com/vanus-labs/vanus-operator/api/restapi/operations/cluster"
	"github.com/vanus-labs/vanus-operator/api/restapi/operations/connector"
	"github.com/vanus-labs/vanus-operator/api/restapi/operations/healthz"
)

//go:generate swagger generate server --target ../../api --name Vanus --spec ../swagger.yaml --principal interface{} --exclude-main

func configureFlags(api *operations.VanusAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.VanusAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	// api.UseRedoc()

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	if api.ClusterCreateClusterHandler == nil {
		api.ClusterCreateClusterHandler = cluster.CreateClusterHandlerFunc(func(params cluster.CreateClusterParams) middleware.Responder {
			return middleware.NotImplemented("operation cluster.CreateCluster has not yet been implemented")
		})
	}
	if api.ConnectorCreateConnectorHandler == nil {
		api.ConnectorCreateConnectorHandler = connector.CreateConnectorHandlerFunc(func(params connector.CreateConnectorParams) middleware.Responder {
			return middleware.NotImplemented("operation connector.CreateConnector has not yet been implemented")
		})
	}
	if api.ClusterDeleteClusterHandler == nil {
		api.ClusterDeleteClusterHandler = cluster.DeleteClusterHandlerFunc(func(params cluster.DeleteClusterParams) middleware.Responder {
			return middleware.NotImplemented("operation cluster.DeleteCluster has not yet been implemented")
		})
	}
	if api.ConnectorDeleteConnectorHandler == nil {
		api.ConnectorDeleteConnectorHandler = connector.DeleteConnectorHandlerFunc(func(params connector.DeleteConnectorParams) middleware.Responder {
			return middleware.NotImplemented("operation connector.DeleteConnector has not yet been implemented")
		})
	}
	if api.ClusterGetClusterHandler == nil {
		api.ClusterGetClusterHandler = cluster.GetClusterHandlerFunc(func(params cluster.GetClusterParams) middleware.Responder {
			return middleware.NotImplemented("operation cluster.GetCluster has not yet been implemented")
		})
	}
	if api.ConnectorGetConnectorHandler == nil {
		api.ConnectorGetConnectorHandler = connector.GetConnectorHandlerFunc(func(params connector.GetConnectorParams) middleware.Responder {
			return middleware.NotImplemented("operation connector.GetConnector has not yet been implemented")
		})
	}
	if api.HealthzHealthzHandler == nil {
		api.HealthzHealthzHandler = healthz.HealthzHandlerFunc(func(params healthz.HealthzParams) middleware.Responder {
			return middleware.NotImplemented("operation healthz.Healthz has not yet been implemented")
		})
	}
	if api.ConnectorListConnectorHandler == nil {
		api.ConnectorListConnectorHandler = connector.ListConnectorHandlerFunc(func(params connector.ListConnectorParams) middleware.Responder {
			return middleware.NotImplemented("operation connector.ListConnector has not yet been implemented")
		})
	}
	if api.ClusterPatchClusterHandler == nil {
		api.ClusterPatchClusterHandler = cluster.PatchClusterHandlerFunc(func(params cluster.PatchClusterParams) middleware.Responder {
			return middleware.NotImplemented("operation cluster.PatchCluster has not yet been implemented")
		})
	}
	if api.ConnectorPatchConnectorHandler == nil {
		api.ConnectorPatchConnectorHandler = connector.PatchConnectorHandlerFunc(func(params connector.PatchConnectorParams) middleware.Responder {
			return middleware.NotImplemented("operation connector.PatchConnector has not yet been implemented")
		})
	}
	if api.ConnectorPatchConnectorsHandler == nil {
		api.ConnectorPatchConnectorsHandler = connector.PatchConnectorsHandlerFunc(func(params connector.PatchConnectorsParams) middleware.Responder {
			return middleware.NotImplemented("operation connector.PatchConnectors has not yet been implemented")
		})
	}

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix".
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation.
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics.
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
