// Code generated by go-swagger; DO NOT EDIT.

package connector

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"
)

// GetConnectorOKCode is the HTTP code returned for type GetConnectorOK
const GetConnectorOKCode int = 200

/*
GetConnectorOK OK

swagger:response getConnectorOK
*/
type GetConnectorOK struct {

	/*
	  In: Body
	*/
	Payload *GetConnectorOKBody `json:"body,omitempty"`
}

// NewGetConnectorOK creates GetConnectorOK with default headers values
func NewGetConnectorOK() *GetConnectorOK {

	return &GetConnectorOK{}
}

// WithPayload adds the payload to the get connector o k response
func (o *GetConnectorOK) WithPayload(payload *GetConnectorOKBody) *GetConnectorOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get connector o k response
func (o *GetConnectorOK) SetPayload(payload *GetConnectorOKBody) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetConnectorOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
