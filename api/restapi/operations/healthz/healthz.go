// Code generated by go-swagger; DO NOT EDIT.

package healthz

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

// HealthzHandlerFunc turns a function with the right signature into a healthz handler
type HealthzHandlerFunc func(HealthzParams) middleware.Responder

// Handle executing the request and returning a response
func (fn HealthzHandlerFunc) Handle(params HealthzParams) middleware.Responder {
	return fn(params)
}

// HealthzHandler interface for that can handle valid healthz params
type HealthzHandler interface {
	Handle(HealthzParams) middleware.Responder
}

// NewHealthz creates a new http.Handler for the healthz operation
func NewHealthz(ctx *middleware.Context, handler HealthzHandler) *Healthz {
	return &Healthz{Context: ctx, Handler: handler}
}

/* Healthz swagger:route GET /healthz/ healthz healthz

healthz

*/
type Healthz struct {
	Context *middleware.Context
	Handler HealthzHandler
}

func (o *Healthz) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewHealthzParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}

// HealthzOKBody healthz o k body
//
// swagger:model HealthzOKBody
type HealthzOKBody struct {

	// code
	// Required: true
	Code *int32 `json:"code"`

	// data
	// Required: true
	Data *models.HealthInfo `json:"data"`

	// message
	// Required: true
	Message *string `json:"message"`
}

// Validate validates this healthz o k body
func (o *HealthzOKBody) Validate(formats strfmt.Registry) error {
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

func (o *HealthzOKBody) validateCode(formats strfmt.Registry) error {

	if err := validate.Required("healthzOK"+"."+"code", "body", o.Code); err != nil {
		return err
	}

	return nil
}

func (o *HealthzOKBody) validateData(formats strfmt.Registry) error {

	if err := validate.Required("healthzOK"+"."+"data", "body", o.Data); err != nil {
		return err
	}

	if o.Data != nil {
		if err := o.Data.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("healthzOK" + "." + "data")
			}
			return err
		}
	}

	return nil
}

func (o *HealthzOKBody) validateMessage(formats strfmt.Registry) error {

	if err := validate.Required("healthzOK"+"."+"message", "body", o.Message); err != nil {
		return err
	}

	return nil
}

// ContextValidate validate this healthz o k body based on the context it is used
func (o *HealthzOKBody) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := o.contextValidateData(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *HealthzOKBody) contextValidateData(ctx context.Context, formats strfmt.Registry) error {

	if o.Data != nil {
		if err := o.Data.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("healthzOK" + "." + "data")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (o *HealthzOKBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *HealthzOKBody) UnmarshalBinary(b []byte) error {
	var res HealthzOKBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}
