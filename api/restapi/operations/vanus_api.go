// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/runtime/security"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	"github.com/linkall-labs/vanus-operator/api/restapi/operations/cluster"
	"github.com/linkall-labs/vanus-operator/api/restapi/operations/connector"
	"github.com/linkall-labs/vanus-operator/api/restapi/operations/healthz"
)

// NewVanusAPI creates a new Vanus instance
func NewVanusAPI(spec *loads.Document) *VanusAPI {
	return &VanusAPI{
		handlers:            make(map[string]map[string]http.Handler),
		formats:             strfmt.Default,
		defaultConsumes:     "application/json",
		defaultProduces:     "application/json",
		customConsumers:     make(map[string]runtime.Consumer),
		customProducers:     make(map[string]runtime.Producer),
		PreServerShutdown:   func() {},
		ServerShutdown:      func() {},
		spec:                spec,
		useSwaggerUI:        false,
		ServeError:          errors.ServeError,
		BasicAuthenticator:  security.BasicAuth,
		APIKeyAuthenticator: security.APIKeyAuth,
		BearerAuthenticator: security.BearerAuth,

		JSONConsumer: runtime.JSONConsumer(),

		JSONProducer: runtime.JSONProducer(),

		ClusterCreateClusterHandler: cluster.CreateClusterHandlerFunc(func(params cluster.CreateClusterParams) middleware.Responder {
			return middleware.NotImplemented("operation cluster.CreateCluster has not yet been implemented")
		}),
		ConnectorCreateConnectorHandler: connector.CreateConnectorHandlerFunc(func(params connector.CreateConnectorParams) middleware.Responder {
			return middleware.NotImplemented("operation connector.CreateConnector has not yet been implemented")
		}),
		ClusterDeleteClusterHandler: cluster.DeleteClusterHandlerFunc(func(params cluster.DeleteClusterParams) middleware.Responder {
			return middleware.NotImplemented("operation cluster.DeleteCluster has not yet been implemented")
		}),
		ConnectorDeleteConnectorHandler: connector.DeleteConnectorHandlerFunc(func(params connector.DeleteConnectorParams) middleware.Responder {
			return middleware.NotImplemented("operation connector.DeleteConnector has not yet been implemented")
		}),
		ClusterGetClusterHandler: cluster.GetClusterHandlerFunc(func(params cluster.GetClusterParams) middleware.Responder {
			return middleware.NotImplemented("operation cluster.GetCluster has not yet been implemented")
		}),
		ConnectorGetConnectorHandler: connector.GetConnectorHandlerFunc(func(params connector.GetConnectorParams) middleware.Responder {
			return middleware.NotImplemented("operation connector.GetConnector has not yet been implemented")
		}),
		HealthzHealthzHandler: healthz.HealthzHandlerFunc(func(params healthz.HealthzParams) middleware.Responder {
			return middleware.NotImplemented("operation healthz.Healthz has not yet been implemented")
		}),
		ConnectorListConnectorHandler: connector.ListConnectorHandlerFunc(func(params connector.ListConnectorParams) middleware.Responder {
			return middleware.NotImplemented("operation connector.ListConnector has not yet been implemented")
		}),
		ClusterPatchClusterHandler: cluster.PatchClusterHandlerFunc(func(params cluster.PatchClusterParams) middleware.Responder {
			return middleware.NotImplemented("operation cluster.PatchCluster has not yet been implemented")
		}),
	}
}

/*VanusAPI this document describes the functions, parameters and return values of the vanus operator API
 */
type VanusAPI struct {
	spec            *loads.Document
	context         *middleware.Context
	handlers        map[string]map[string]http.Handler
	formats         strfmt.Registry
	customConsumers map[string]runtime.Consumer
	customProducers map[string]runtime.Producer
	defaultConsumes string
	defaultProduces string
	Middleware      func(middleware.Builder) http.Handler
	useSwaggerUI    bool

	// BasicAuthenticator generates a runtime.Authenticator from the supplied basic auth function.
	// It has a default implementation in the security package, however you can replace it for your particular usage.
	BasicAuthenticator func(security.UserPassAuthentication) runtime.Authenticator

	// APIKeyAuthenticator generates a runtime.Authenticator from the supplied token auth function.
	// It has a default implementation in the security package, however you can replace it for your particular usage.
	APIKeyAuthenticator func(string, string, security.TokenAuthentication) runtime.Authenticator

	// BearerAuthenticator generates a runtime.Authenticator from the supplied bearer token auth function.
	// It has a default implementation in the security package, however you can replace it for your particular usage.
	BearerAuthenticator func(string, security.ScopedTokenAuthentication) runtime.Authenticator

	// JSONConsumer registers a consumer for the following mime types:
	//   - application/json
	JSONConsumer runtime.Consumer

	// JSONProducer registers a producer for the following mime types:
	//   - application/json
	JSONProducer runtime.Producer

	// ClusterCreateClusterHandler sets the operation handler for the create cluster operation
	ClusterCreateClusterHandler cluster.CreateClusterHandler
	// ConnectorCreateConnectorHandler sets the operation handler for the create connector operation
	ConnectorCreateConnectorHandler connector.CreateConnectorHandler
	// ClusterDeleteClusterHandler sets the operation handler for the delete cluster operation
	ClusterDeleteClusterHandler cluster.DeleteClusterHandler
	// ConnectorDeleteConnectorHandler sets the operation handler for the delete connector operation
	ConnectorDeleteConnectorHandler connector.DeleteConnectorHandler
	// ClusterGetClusterHandler sets the operation handler for the get cluster operation
	ClusterGetClusterHandler cluster.GetClusterHandler
	// ConnectorGetConnectorHandler sets the operation handler for the get connector operation
	ConnectorGetConnectorHandler connector.GetConnectorHandler
	// HealthzHealthzHandler sets the operation handler for the healthz operation
	HealthzHealthzHandler healthz.HealthzHandler
	// ConnectorListConnectorHandler sets the operation handler for the list connector operation
	ConnectorListConnectorHandler connector.ListConnectorHandler
	// ClusterPatchClusterHandler sets the operation handler for the patch cluster operation
	ClusterPatchClusterHandler cluster.PatchClusterHandler

	// ServeError is called when an error is received, there is a default handler
	// but you can set your own with this
	ServeError func(http.ResponseWriter, *http.Request, error)

	// PreServerShutdown is called before the HTTP(S) server is shutdown
	// This allows for custom functions to get executed before the HTTP(S) server stops accepting traffic
	PreServerShutdown func()

	// ServerShutdown is called when the HTTP(S) server is shut down and done
	// handling all active connections and does not accept connections any more
	ServerShutdown func()

	// Custom command line argument groups with their descriptions
	CommandLineOptionsGroups []swag.CommandLineOptionsGroup

	// User defined logger function.
	Logger func(string, ...interface{})
}

// UseRedoc for documentation at /docs
func (o *VanusAPI) UseRedoc() {
	o.useSwaggerUI = false
}

// UseSwaggerUI for documentation at /docs
func (o *VanusAPI) UseSwaggerUI() {
	o.useSwaggerUI = true
}

// SetDefaultProduces sets the default produces media type
func (o *VanusAPI) SetDefaultProduces(mediaType string) {
	o.defaultProduces = mediaType
}

// SetDefaultConsumes returns the default consumes media type
func (o *VanusAPI) SetDefaultConsumes(mediaType string) {
	o.defaultConsumes = mediaType
}

// SetSpec sets a spec that will be served for the clients.
func (o *VanusAPI) SetSpec(spec *loads.Document) {
	o.spec = spec
}

// DefaultProduces returns the default produces media type
func (o *VanusAPI) DefaultProduces() string {
	return o.defaultProduces
}

// DefaultConsumes returns the default consumes media type
func (o *VanusAPI) DefaultConsumes() string {
	return o.defaultConsumes
}

// Formats returns the registered string formats
func (o *VanusAPI) Formats() strfmt.Registry {
	return o.formats
}

// RegisterFormat registers a custom format validator
func (o *VanusAPI) RegisterFormat(name string, format strfmt.Format, validator strfmt.Validator) {
	o.formats.Add(name, format, validator)
}

// Validate validates the registrations in the VanusAPI
func (o *VanusAPI) Validate() error {
	var unregistered []string

	if o.JSONConsumer == nil {
		unregistered = append(unregistered, "JSONConsumer")
	}

	if o.JSONProducer == nil {
		unregistered = append(unregistered, "JSONProducer")
	}

	if o.ClusterCreateClusterHandler == nil {
		unregistered = append(unregistered, "cluster.CreateClusterHandler")
	}
	if o.ConnectorCreateConnectorHandler == nil {
		unregistered = append(unregistered, "connector.CreateConnectorHandler")
	}
	if o.ClusterDeleteClusterHandler == nil {
		unregistered = append(unregistered, "cluster.DeleteClusterHandler")
	}
	if o.ConnectorDeleteConnectorHandler == nil {
		unregistered = append(unregistered, "connector.DeleteConnectorHandler")
	}
	if o.ClusterGetClusterHandler == nil {
		unregistered = append(unregistered, "cluster.GetClusterHandler")
	}
	if o.ConnectorGetConnectorHandler == nil {
		unregistered = append(unregistered, "connector.GetConnectorHandler")
	}
	if o.HealthzHealthzHandler == nil {
		unregistered = append(unregistered, "healthz.HealthzHandler")
	}
	if o.ConnectorListConnectorHandler == nil {
		unregistered = append(unregistered, "connector.ListConnectorHandler")
	}
	if o.ClusterPatchClusterHandler == nil {
		unregistered = append(unregistered, "cluster.PatchClusterHandler")
	}

	if len(unregistered) > 0 {
		return fmt.Errorf("missing registration: %s", strings.Join(unregistered, ", "))
	}

	return nil
}

// ServeErrorFor gets a error handler for a given operation id
func (o *VanusAPI) ServeErrorFor(operationID string) func(http.ResponseWriter, *http.Request, error) {
	return o.ServeError
}

// AuthenticatorsFor gets the authenticators for the specified security schemes
func (o *VanusAPI) AuthenticatorsFor(schemes map[string]spec.SecurityScheme) map[string]runtime.Authenticator {
	return nil
}

// Authorizer returns the registered authorizer
func (o *VanusAPI) Authorizer() runtime.Authorizer {
	return nil
}

// ConsumersFor gets the consumers for the specified media types.
// MIME type parameters are ignored here.
func (o *VanusAPI) ConsumersFor(mediaTypes []string) map[string]runtime.Consumer {
	result := make(map[string]runtime.Consumer, len(mediaTypes))
	for _, mt := range mediaTypes {
		switch mt {
		case "application/json":
			result["application/json"] = o.JSONConsumer
		}

		if c, ok := o.customConsumers[mt]; ok {
			result[mt] = c
		}
	}
	return result
}

// ProducersFor gets the producers for the specified media types.
// MIME type parameters are ignored here.
func (o *VanusAPI) ProducersFor(mediaTypes []string) map[string]runtime.Producer {
	result := make(map[string]runtime.Producer, len(mediaTypes))
	for _, mt := range mediaTypes {
		switch mt {
		case "application/json":
			result["application/json"] = o.JSONProducer
		}

		if p, ok := o.customProducers[mt]; ok {
			result[mt] = p
		}
	}
	return result
}

// HandlerFor gets a http.Handler for the provided operation method and path
func (o *VanusAPI) HandlerFor(method, path string) (http.Handler, bool) {
	if o.handlers == nil {
		return nil, false
	}
	um := strings.ToUpper(method)
	if _, ok := o.handlers[um]; !ok {
		return nil, false
	}
	if path == "/" {
		path = ""
	}
	h, ok := o.handlers[um][path]
	return h, ok
}

// Context returns the middleware context for the vanus API
func (o *VanusAPI) Context() *middleware.Context {
	if o.context == nil {
		o.context = middleware.NewRoutableContext(o.spec, o, nil)
	}

	return o.context
}

func (o *VanusAPI) initHandlerCache() {
	o.Context() // don't care about the result, just that the initialization happened
	if o.handlers == nil {
		o.handlers = make(map[string]map[string]http.Handler)
	}

	if o.handlers["POST"] == nil {
		o.handlers["POST"] = make(map[string]http.Handler)
	}
	o.handlers["POST"]["/cluster"] = cluster.NewCreateCluster(o.context, o.ClusterCreateClusterHandler)
	if o.handlers["POST"] == nil {
		o.handlers["POST"] = make(map[string]http.Handler)
	}
	o.handlers["POST"]["/connector"] = connector.NewCreateConnector(o.context, o.ConnectorCreateConnectorHandler)
	if o.handlers["DELETE"] == nil {
		o.handlers["DELETE"] = make(map[string]http.Handler)
	}
	o.handlers["DELETE"]["/cluster"] = cluster.NewDeleteCluster(o.context, o.ClusterDeleteClusterHandler)
	if o.handlers["DELETE"] == nil {
		o.handlers["DELETE"] = make(map[string]http.Handler)
	}
	o.handlers["DELETE"]["/connector"] = connector.NewDeleteConnector(o.context, o.ConnectorDeleteConnectorHandler)
	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/cluster"] = cluster.NewGetCluster(o.context, o.ClusterGetClusterHandler)
	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/connector"] = connector.NewGetConnector(o.context, o.ConnectorGetConnectorHandler)
	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/healthz"] = healthz.NewHealthz(o.context, o.HealthzHealthzHandler)
	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/connectors"] = connector.NewListConnector(o.context, o.ConnectorListConnectorHandler)
	if o.handlers["PATCH"] == nil {
		o.handlers["PATCH"] = make(map[string]http.Handler)
	}
	o.handlers["PATCH"]["/cluster"] = cluster.NewPatchCluster(o.context, o.ClusterPatchClusterHandler)
}

// Serve creates a http handler to serve the API over HTTP
// can be used directly in http.ListenAndServe(":8000", api.Serve(nil))
func (o *VanusAPI) Serve(builder middleware.Builder) http.Handler {
	o.Init()

	if o.Middleware != nil {
		return o.Middleware(builder)
	}
	if o.useSwaggerUI {
		return o.context.APIHandlerSwaggerUI(builder)
	}
	return o.context.APIHandler(builder)
}

// Init allows you to just initialize the handler cache, you can then recompose the middleware as you see fit
func (o *VanusAPI) Init() {
	if len(o.handlers) == 0 {
		o.initHandlerCache()
	}
}

// RegisterConsumer allows you to add (or override) a consumer for a media type.
func (o *VanusAPI) RegisterConsumer(mediaType string, consumer runtime.Consumer) {
	o.customConsumers[mediaType] = consumer
}

// RegisterProducer allows you to add (or override) a producer for a media type.
func (o *VanusAPI) RegisterProducer(mediaType string, producer runtime.Producer) {
	o.customProducers[mediaType] = producer
}

// AddMiddlewareFor adds a http middleware to existing handler
func (o *VanusAPI) AddMiddlewareFor(method, path string, builder middleware.Builder) {
	um := strings.ToUpper(method)
	if path == "/" {
		path = ""
	}
	o.Init()
	if h, ok := o.handlers[um][path]; ok {
		o.handlers[method][path] = builder(h)
	}
}
