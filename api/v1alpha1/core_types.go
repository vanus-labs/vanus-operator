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

// CoreSpec defines the desired state of Core
type CoreSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Replicas is the number of nodes in the Vanus cluster components.
	Replicas Replicas `json:"replicas,omitempty"`
	// Replicas is the Vanus cluster version. All components remain the same version.
	Version string `json:"version,omitempty"`
	// ImagePullPolicy defines how the image is pulled
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// The desired compute resource requirements of Pods in the cluster.
	// +kubebuilder:default:={limits: {cpu: "500m", memory: "1024Mi"}, requests: {cpu: "250m", memory: "512Mi"}}
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// VolumeClaimTemplates is a list of claims that pods are allowed to reference.
	// The StatefulSet controller is responsible for mapping network identities to
	// claims in a way that maintains the identity of a pod. Every claim in
	// this list must have at least one matching (by name) volumeMount in one
	// container in the template. A claim in this list takes precedence over
	// any volumes in the template, with the same name.
	// +optional
	VolumeClaimTemplates []corev1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`
}

// CoreStatus defines the observed state of Core
type CoreStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Core is the Schema for the cores API
type Core struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CoreSpec   `json:"spec,omitempty"`
	Status CoreStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CoreList contains a list of Core
type CoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Core `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Core{}, &CoreList{})
}

type Replicas struct {
	// Replicas is the number of nodes in the Controller. Each node is deployed as a Replica in a StatefulSet.
	// This value should be an odd number to ensure the resultant cluster can establish exactly one quorum of nodes
	// in the event of a fragmenting network partition.
	// +optional
	// +kubebuilder:validation:Minimum:=0
	// +kubebuilder:default:=3
	Controller int32 `json:"controller,omitempty"`
	// Replicas is the number of nodes in the Store. Each node is deployed as a Replica in a StatefulSet.
	// This value should be an odd number to ensure the resultant cluster can establish exactly one quorum of nodes
	// in the event of a fragmenting network partition.
	// +optional
	// +kubebuilder:validation:Minimum:=0
	// +kubebuilder:default:=3
	Store int32 `json:"store,omitempty"`
	// Replicas is the number of nodes in the Trigger. Each node is deployed as a Replica in a Deployment.
	// +optional
	// +kubebuilder:validation:Minimum:=0
	// +kubebuilder:default:=1
	Trigger int32 `json:"trigger,omitempty"`
	// Replicas is the number of nodes in the Timer. Each node is deployed as a Replica in a Deployment.
	// This value should be greater than 1, because Timer uses the active and standby architecture to ensure
	// that the master fails and the slave quickly takes over the business.
	// +optional
	// +kubebuilder:validation:Minimum:=0
	// +kubebuilder:default:=2
	Timer int32 `json:"timer,omitempty"`
	// Replicas is the number of nodes in the Gateway. Each node is deployed as a Replica in a Deployment.
	// +optional
	// +kubebuilder:validation:Minimum:=0
	// +kubebuilder:default:=1
	Gateway int32 `json:"gateway,omitempty"`
}
