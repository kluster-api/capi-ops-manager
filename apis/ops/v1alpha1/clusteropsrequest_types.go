/*
Copyright 2024.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
)

// ClusterOpsRequestSpec defines the desired state of ClusterOpsRequest
type ClusterOpsRequestSpec struct {
	ClusterRef kmapi.ObjectReference `json:"clusterRef"`
	Type       ClusterOpsRequestType `json:"type"`
	// +optional
	UpdateVersion *ClusterUpdateVersionSpec `json:"updateVersion,omitempty"`
	// ApplyOption is to control the execution of OpsRequest depending on the database state.
}

type ClusterOpsRequestType string

// RedisOpsRequestTypeUpdateVersion is a RedisOpsRequestType of type UpdateVersion.
// +kubebuilder:validation:Enum=UpdateVersion
const ClusterOpsRequestTypeUpdateVersion ClusterOpsRequestType = "UpdateVersion"

type ClusterUpdateVersionSpec struct {
	TargetVersion TargetVersion `json:"targetVersion,omitempty"`
}

type TargetVersion struct {
	Cluster   *string           `json:"cluster,omitempty"`
	Providers *ProviderVersions `json:"providers,omitempty"`
}

type ProviderVersions struct {
	Core           string `json:"core,omitempty"`
	Bootstrap      string `json:"bootstrap,omitempty"`
	ControlPlane   string `json:"controlPlane,omitempty"`
	Infrastructure string `json:"infrastructure,omitempty"`
}

// ClusterOpsRequestStatus defines the observed state of ClusterOpsRequest
type ClusterOpsRequestStatus struct {
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	Conditions []kmapi.Condition `json:"conditions"`
	// +optional
	Phase ClusterOpsRequestPhase `json:"phase"`
}

// ClusterOpsRequest is the Schema for the clusteropsrequests API

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type ClusterOpsRequest struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec ClusterOpsRequestSpec `json:"spec,omitempty"`
	// +optional
	Status ClusterOpsRequestStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterOpsRequestList contains a list of ClusterOpsRequest
type ClusterOpsRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterOpsRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterOpsRequest{}, &ClusterOpsRequestList{})
}
