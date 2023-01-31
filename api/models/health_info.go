// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// HealthInfo operator info
//
// swagger:model Health_info
type HealthInfo struct {

	// operator status
	Status string `json:"status,omitempty"`
}

// Validate validates this health info
func (m *HealthInfo) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this health info based on context it is used
func (m *HealthInfo) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *HealthInfo) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *HealthInfo) UnmarshalBinary(b []byte) error {
	var res HealthInfo
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
