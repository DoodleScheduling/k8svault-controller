/*


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

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VaultBindingSpec defines the desired state of VaultBinding
type VaultBindingSpec struct {
	*VaultSpec `json:",inline"`

	// Define the secrets which must be mapped to vault
	// +optional
	Fields []FieldMapping `json:"fields,omitempty"`

	// By default existing matching fields in vault do not get overwritten
	// +optional
	ForceApply bool `json:"forceApply,omitempty"`

	// The kubernetes secret the VaultBinding is referring to
	// +required
	Secret *corev1.SecretReference `json:"secret"`
}

func (in *VaultBindingSpec) IsForceApply() bool {
	return in.ForceApply
}

func (in *VaultBindingSpec) GetPath() string {
	return in.Path
}

func (in *VaultBindingSpec) GetFieldMapping() []FieldMapping {
	return in.Fields
}

// VaultBindingStatus defines the observed state of VaultBinding
type VaultBindingStatus struct {
	// Conditions holds the conditions for the VaultBinding.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the last generation reconciled by the controller
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Vault Status (not implemented yet)
	Vault VaultBindingVaultStatus `json:",inline"`
}

// VaultBindingNotBound de
func VaultBindingNotBound(binding VaultBinding, reason, message string) VaultBinding {
	setResourceCondition(&binding, BoundCondition, metav1.ConditionFalse, reason, message)
	return binding
}

// VaultBindingBound de
func VaultBindingBound(binding VaultBinding, reason, message string) VaultBinding {
	setResourceCondition(&binding, BoundCondition, metav1.ConditionTrue, reason, message)
	return binding
}

// GetStatusConditions returns a pointer to the Status.Conditions slice
func (in *VaultBinding) GetStatusConditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

type VaultBindingVaultStatus struct {
	Address string `json:"address,omitempty"`
	Path    string `json:"path,omitempty"`
	Fields  string `json:"fields,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=vb
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Bound\")].status",description=""
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Bound\")].message",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""

// VaultBinding is the Schema for the vaultbindings API
type VaultBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VaultBindingSpec   `json:"spec,omitempty"`
	Status VaultBindingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VaultBindingList contains a list of VaultBinding
type VaultBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VaultBinding{}, &VaultBindingList{})
}
