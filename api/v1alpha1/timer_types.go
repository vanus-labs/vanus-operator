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

// TimerSpec defines the desired state of Timer
type TimerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Replicas is the number of nodes in the Timer. Each node is deployed as a Replica in a Deployment.
	// This value should be greater than 1, because Timer uses the active and standby architecture to ensure
	// that the master fails and the slave quickly takes over the business.
	// +optional
	// +kubebuilder:validation:Minimum:=0
	// +kubebuilder:default:=2
	Replicas *int32 `json:"replicas,omitempty"`
	// Image is the name of the gateway docker image to use for the Pods.
	// Must be provided together with ImagePullSecrets in order to use an image in a private registry.
	Image string `json:"image,omitempty"`
	// ImagePullPolicy defines how the image is pulled
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// The desired compute resource requirements of Pods in the cluster.
	// +kubebuilder:default:={limits: {cpu: "500m", memory: "1024Mi"}, requests: {cpu: "250m", memory: "512Mi"}}
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// TimerStatus defines the observed state of Timer
type TimerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Timer is the Schema for the timers API
type Timer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TimerSpec   `json:"spec,omitempty"`
	Status TimerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TimerList contains a list of Timer
type TimerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Timer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Timer{}, &TimerList{})
}
