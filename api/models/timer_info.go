// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// TimerInfo timer info
//
// swagger:model Timer_info
type TimerInfo struct {

	// timer replicas
	Replicas int32 `json:"replicas,omitempty"`

	// timer version
	Version string `json:"version,omitempty"`
}

// Validate validates this timer info
func (m *TimerInfo) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this timer info based on context it is used
func (m *TimerInfo) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *TimerInfo) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *TimerInfo) UnmarshalBinary(b []byte) error {
	var res TimerInfo
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}