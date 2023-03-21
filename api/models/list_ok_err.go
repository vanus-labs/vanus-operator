// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ListOkErr list ok err
//
// swagger:model ListOkErr
type ListOkErr struct {

	// failed
	Failed []*ListOkErrFailedItems0 `json:"failed"`

	// successed
	Successed []string `json:"successed"`
}

// Validate validates this list ok err
func (m *ListOkErr) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateFailed(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ListOkErr) validateFailed(formats strfmt.Registry) error {
	if swag.IsZero(m.Failed) { // not required
		return nil
	}

	for i := 0; i < len(m.Failed); i++ {
		if swag.IsZero(m.Failed[i]) { // not required
			continue
		}

		if m.Failed[i] != nil {
			if err := m.Failed[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("failed" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// ContextValidate validate this list ok err based on the context it is used
func (m *ListOkErr) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateFailed(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ListOkErr) contextValidateFailed(ctx context.Context, formats strfmt.Registry) error {

	for i := 0; i < len(m.Failed); i++ {

		if m.Failed[i] != nil {
			if err := m.Failed[i].ContextValidate(ctx, formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("failed" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *ListOkErr) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ListOkErr) UnmarshalBinary(b []byte) error {
	var res ListOkErr
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

// ListOkErrFailedItems0 list ok err failed items0
//
// swagger:model ListOkErrFailedItems0
type ListOkErrFailedItems0 struct {

	// name
	Name string `json:"name,omitempty"`

	// reason
	Reason string `json:"reason,omitempty"`
}

// Validate validates this list ok err failed items0
func (m *ListOkErrFailedItems0) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this list ok err failed items0 based on context it is used
func (m *ListOkErrFailedItems0) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *ListOkErrFailedItems0) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ListOkErrFailedItems0) UnmarshalBinary(b []byte) error {
	var res ListOkErrFailedItems0
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}