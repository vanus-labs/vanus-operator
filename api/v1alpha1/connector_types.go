/*
Copyright 2023 Linkall Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConnectorSpec defines the desired state of Connector
type ConnectorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Kind is the kind of connector, support source/sink.
	Kind string `json:"kind,omitempty"`
	// Name is the name of connector.
	Name string `json:"name,omitempty"`
	// Type is the type of connector.
	Type string `json:"type,omitempty"`
	// Config is the file of config.
	Config string `json:"config,omitempty"`
	// Image is the name of the controller docker image to use for the Pods.
	// Must be provided together with ImagePullSecrets in order to use an image in a private registry.
	Image string `json:"image,omitempty"`
	// ImagePullPolicy defines how the image is pulled
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
}

// ConnectorStatus defines the observed state of Connector
type ConnectorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Connector is the Schema for the connectors API
type Connector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConnectorSpec   `json:"spec,omitempty"`
	Status ConnectorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ConnectorList contains a list of Connector
type ConnectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Connector `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Connector{}, &ConnectorList{})
}
