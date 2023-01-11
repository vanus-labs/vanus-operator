// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ConnectorInfo Connector info
//
// swagger:model Connector_info
type ConnectorInfo struct {

	// connector type
	Kind string `json:"kind,omitempty"`

	// connector name
	Name string `json:"name,omitempty"`
}

// Validate validates this connector info
func (m *ConnectorInfo) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this connector info based on context it is used
func (m *ConnectorInfo) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *ConnectorInfo) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ConnectorInfo) UnmarshalBinary(b []byte) error {
	var res ConnectorInfo
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
