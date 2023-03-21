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

	"github.com/vanus-labs/vanus-operator/api/models"
)

// PatchConnectorsHandlerFunc turns a function with the right signature into a patch connectors handler
type PatchConnectorsHandlerFunc func(PatchConnectorsParams) middleware.Responder

// Handle executing the request and returning a response
func (fn PatchConnectorsHandlerFunc) Handle(params PatchConnectorsParams) middleware.Responder {
	return fn(params)
}

// PatchConnectorsHandler interface for that can handle valid patch connectors params
type PatchConnectorsHandler interface {
	Handle(PatchConnectorsParams) middleware.Responder
}

// NewPatchConnectors creates a new http.Handler for the patch connectors operation
func NewPatchConnectors(ctx *middleware.Context, handler PatchConnectorsHandler) *PatchConnectors {
	return &PatchConnectors{Context: ctx, Handler: handler}
}

/*
	PatchConnectors swagger:route PATCH /connectors/ connector patchConnectors

patch Connectors
*/
type PatchConnectors struct {
	Context *middleware.Context
	Handler PatchConnectorsHandler
}

func (o *PatchConnectors) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewPatchConnectorsParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}

// PatchConnectorsOKBody patch connectors o k body
//
// swagger:model PatchConnectorsOKBody
type PatchConnectorsOKBody struct {

	// code
	// Required: true
	Code *int32 `json:"code"`

	// data
	// Required: true
	Data *models.ListOkErr `json:"data"`

	// message
	// Required: true
	Message *string `json:"message"`
}

// Validate validates this patch connectors o k body
func (o *PatchConnectorsOKBody) Validate(formats strfmt.Registry) error {
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

func (o *PatchConnectorsOKBody) validateCode(formats strfmt.Registry) error {

	if err := validate.Required("patchConnectorsOK"+"."+"code", "body", o.Code); err != nil {
		return err
	}

	return nil
}

func (o *PatchConnectorsOKBody) validateData(formats strfmt.Registry) error {

	if err := validate.Required("patchConnectorsOK"+"."+"data", "body", o.Data); err != nil {
		return err
	}

	if o.Data != nil {
		if err := o.Data.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("patchConnectorsOK" + "." + "data")
			}
			return err
		}
	}

	return nil
}

func (o *PatchConnectorsOKBody) validateMessage(formats strfmt.Registry) error {

	if err := validate.Required("patchConnectorsOK"+"."+"message", "body", o.Message); err != nil {
		return err
	}

	return nil
}

// ContextValidate validate this patch connectors o k body based on the context it is used
func (o *PatchConnectorsOKBody) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := o.contextValidateData(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *PatchConnectorsOKBody) contextValidateData(ctx context.Context, formats strfmt.Registry) error {

	if o.Data != nil {
		if err := o.Data.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("patchConnectorsOK" + "." + "data")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (o *PatchConnectorsOKBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *PatchConnectorsOKBody) UnmarshalBinary(b []byte) error {
	var res PatchConnectorsOKBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}