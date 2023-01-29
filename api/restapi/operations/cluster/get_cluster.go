// Code generated by go-swagger; DO NOT EDIT.

package cluster

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

// GetClusterHandlerFunc turns a function with the right signature into a get cluster handler
type GetClusterHandlerFunc func(GetClusterParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetClusterHandlerFunc) Handle(params GetClusterParams) middleware.Responder {
	return fn(params)
}

// GetClusterHandler interface for that can handle valid get cluster params
type GetClusterHandler interface {
	Handle(GetClusterParams) middleware.Responder
}

// NewGetCluster creates a new http.Handler for the get cluster operation
func NewGetCluster(ctx *middleware.Context, handler GetClusterHandler) *GetCluster {
	return &GetCluster{Context: ctx, Handler: handler}
}

/* GetCluster swagger:route GET /cluster/ cluster getCluster

get Cluster

*/
type GetCluster struct {
	Context *middleware.Context
	Handler GetClusterHandler
}

func (o *GetCluster) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewGetClusterParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}

// GetClusterOKBody get cluster o k body
//
// swagger:model GetClusterOKBody
type GetClusterOKBody struct {

	// code
	// Required: true
	Code *int32 `json:"code"`

	// data
	// Required: true
	Data *models.ClusterInfo `json:"data"`

	// message
	// Required: true
	Message *string `json:"message"`
}

// Validate validates this get cluster o k body
func (o *GetClusterOKBody) Validate(formats strfmt.Registry) error {
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

func (o *GetClusterOKBody) validateCode(formats strfmt.Registry) error {

	if err := validate.Required("getClusterOK"+"."+"code", "body", o.Code); err != nil {
		return err
	}

	return nil
}

func (o *GetClusterOKBody) validateData(formats strfmt.Registry) error {

	if err := validate.Required("getClusterOK"+"."+"data", "body", o.Data); err != nil {
		return err
	}

	if o.Data != nil {
		if err := o.Data.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("getClusterOK" + "." + "data")
			}
			return err
		}
	}

	return nil
}

func (o *GetClusterOKBody) validateMessage(formats strfmt.Registry) error {

	if err := validate.Required("getClusterOK"+"."+"message", "body", o.Message); err != nil {
		return err
	}

	return nil
}

// ContextValidate validate this get cluster o k body based on the context it is used
func (o *GetClusterOKBody) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := o.contextValidateData(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *GetClusterOKBody) contextValidateData(ctx context.Context, formats strfmt.Registry) error {

	if o.Data != nil {
		if err := o.Data.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("getClusterOK" + "." + "data")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (o *GetClusterOKBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *GetClusterOKBody) UnmarshalBinary(b []byte) error {
	var res GetClusterOKBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}