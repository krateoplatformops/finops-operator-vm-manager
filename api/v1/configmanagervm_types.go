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
// +kubebuilder:object:generate=true
package v1

import (
	"github.com/krateoplatformops/finops-operator-vm-manager/providers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigManagerVMSpec struct {
	ResourceProvider          string                    `json:"resourceProvider"`
	ProviderSpecificResources ProviderSpecificResources `json:"providerSpecificResources,omitempty"`
}

type ProviderSpecificResources struct {
	// +optional
	AzureLogin providers.Azure `json:"azure,omitempty"`
}

// ConfigManagerVMStatus defines the observed state of ConfigManagerVM
type ConfigManagerVMStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ConfigManagerVM is the Schema for the configmanagervms API
type ConfigManagerVM struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigManagerVMSpec   `json:"spec,omitempty"`
	Status ConfigManagerVMStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ConfigManagerVMList contains a list of ConfigManagerVM
type ConfigManagerVMList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConfigManagerVM `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConfigManagerVM{}, &ConfigManagerVMList{})
}
