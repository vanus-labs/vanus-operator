// Code generated by go-swagger; DO NOT EDIT.

package connector

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"io"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/vanus-labs/vanus-operator/api/models"
)

// NewPatchConnectorsParams creates a new PatchConnectorsParams object
//
// There are no default values defined in the spec.
func NewPatchConnectorsParams() PatchConnectorsParams {

	return PatchConnectorsParams{}
}

// PatchConnectorsParams contains all the bound params for the patch connectors operation
// typically these are obtained from a http.Request
//
// swagger:parameters patchConnectors
type PatchConnectorsParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*patch info
	  Required: true
	  In: body
	*/
	Connectors []*models.ConnectorPatch
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewPatchConnectorsParams() beforehand.
func (o *PatchConnectorsParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body []*models.ConnectorPatch
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			if err == io.EOF {
				res = append(res, errors.Required("connectors", "body", ""))
			} else {
				res = append(res, errors.NewParseError("connectors", "body", "", err))
			}
		} else {

			// validate array of body objects
			for i := range body {
				if body[i] == nil {
					continue
				}
				if err := body[i].Validate(route.Formats); err != nil {
					res = append(res, err)
					break
				}
			}

			if len(res) == 0 {
				o.Connectors = body
			}
		}
	} else {
		res = append(res, errors.Required("connectors", "body", ""))
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
