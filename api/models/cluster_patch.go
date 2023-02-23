// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ClusterPatch cluster patch params
//
// swagger:model Cluster_patch
type ClusterPatch struct {

	// controller replicas
	ControllerReplicas int32 `json:"controller_replicas,omitempty"`

	// gateway replicas
	GatewayReplicas int32 `json:"gateway_replicas,omitempty"`

	// store replicas
	StoreReplicas int32 `json:"store_replicas,omitempty"`

	// timer replicas
	TimerReplicas int32 `json:"timer_replicas,omitempty"`

	// trigger replicas
	TriggerReplicas int32 `json:"trigger_replicas,omitempty"`

	// cluster version
	Version string `json:"version,omitempty"`
}

// Validate validates this cluster patch
func (m *ClusterPatch) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this cluster patch based on context it is used
func (m *ClusterPatch) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *ClusterPatch) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ClusterPatch) UnmarshalBinary(b []byte) error {
	var res ClusterPatch
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
