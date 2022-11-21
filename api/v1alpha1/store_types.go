/*
Copyright 2022.

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

// StoreSpec defines the desired state of Store
type StoreSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Replicas is the number of nodes in the Store. Each node is deployed as a Replica in a StatefulSet.
	// This value should be an odd number to ensure the resultant cluster can establish exactly one quorum of nodes
	// in the event of a fragmenting network partition.
	// +optional
	// +kubebuilder:validation:Minimum:=0
	// +kubebuilder:default:=3
	Replicas *int32 `json:"replicas,omitempty"`
	// Image is the name of the store docker image to use for the Pods.
	// Must be provided together with ImagePullSecrets in order to use an image in a private registry.
	Image string `json:"image,omitempty"`
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

// StoreStatus defines the observed state of Store
type StoreStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Store is the Schema for the stores API
type Store struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StoreSpec   `json:"spec,omitempty"`
	Status StoreStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StoreList contains a list of Store
type StoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Store `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Store{}, &StoreList{})
}
