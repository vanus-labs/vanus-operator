// Code generated by go-swagger; DO NOT EDIT.

package connector

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"context"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"

	"github.com/linkall-labs/vanus-operator/api/models"
)

// CreateConnectorHandlerFunc turns a function with the right signature into a create connector handler
type CreateConnectorHandlerFunc func(CreateConnectorParams) middleware.Responder

// Handle executing the request and returning a response
func (fn CreateConnectorHandlerFunc) Handle(params CreateConnectorParams) middleware.Responder {
	return fn(params)
}

// CreateConnectorHandler interface for that can handle valid create connector params
type CreateConnectorHandler interface {
	Handle(CreateConnectorParams) middleware.Responder
}

// NewCreateConnector creates a new http.Handler for the create connector operation
func NewCreateConnector(ctx *middleware.Context, handler CreateConnectorHandler) *CreateConnector {
	return &CreateConnector{Context: ctx, Handler: handler}
}

/* CreateConnector swagger:route POST /connectors/ connector createConnector

create Connector

*/
type CreateConnector struct {
	Context *middleware.Context
	Handler CreateConnectorHandler
}

func (o *CreateConnector) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewCreateConnectorParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}

// CreateConnectorOKBody create connector o k body
//
// swagger:model CreateConnectorOKBody
type CreateConnectorOKBody struct {

	// code
	// Required: true
	Code *int32 `json:"code"`

	// data
	// Required: true
	Data *models.ConnectorInfo `json:"data"`

	// message
	// Required: true
	Message *string `json:"message"`
}

// Validate validates this create connector o k body
func (o *CreateConnectorOKBody) Validate(formats strfmt.Registry) error {
	var res []error

	if err := o.validateCode(formats); err != nil {
		res = append(res, err)
	}

	if err := o.validateData(formats); err != nil {
		res = append(res, err)
	}

	if err := o.validateMessage(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *CreateConnectorOKBody) validateCode(formats strfmt.Registry) error {

	if err := validate.Required("createConnectorOK"+"."+"code", "body", o.Code); err != nil {
		return err
	}

	return nil
}

func (o *CreateConnectorOKBody) validateData(formats strfmt.Registry) error {

	if err := validate.Required("createConnectorOK"+"."+"data", "body", o.Data); err != nil {
		return err
	}

	if o.Data != nil {
		if err := o.Data.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("createConnectorOK" + "." + "data")
			}
			return err
		}
	}

	return nil
}

func (o *CreateConnectorOKBody) validateMessage(formats strfmt.Registry) error {

	if err := validate.Required("createConnectorOK"+"."+"message", "body", o.Message); err != nil {
		return err
	}

	return nil
}

// ContextValidate validate this create connector o k body based on the context it is used
func (o *CreateConnectorOKBody) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := o.contextValidateData(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *CreateConnectorOKBody) contextValidateData(ctx context.Context, formats strfmt.Registry) error {

	if o.Data != nil {
		if err := o.Data.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("createConnectorOK" + "." + "data")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (o *CreateConnectorOKBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *CreateConnectorOKBody) UnmarshalBinary(b []byte) error {
	var res CreateConnectorOKBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}
