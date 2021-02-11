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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VaultMirrorSpec defines the desired state of VaultMirror
type VaultMirrorSpec struct {
	// Source vault server to mirror
	// +required
	Source *VaultSpec `json:"source"`

	// Destination vault server
	// +required
	Destination *VaultSpec `json:"destination"`

	// By default existing matching fields in vault do not get overwritten
	// +optional
	ForceApply bool `json:"forceApply"`

	// Define the secrets which must be mapped to vault
	// +optional
	Fields []FieldMapping `json:"fields"`
}

// VaultMirrorStatus defines the observed state of VaultMirror
type VaultMirrorStatus struct {
	// Conditions holds the conditions for the VaultMirror.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Vault Status (not implemented yet)
	Vault VaultMirrorVaultStatus `json:",inline"`
}

func (in *VaultMirrorSpec) IsForceApply() bool {
	return in.ForceApply
}

func (in *VaultMirrorSpec) GetPath() string {
	return in.Destination.Path
}

func (in *VaultMirrorSpec) GetFieldMapping() []FieldMapping {
	return in.Fields
}

// VaultMirrorNotBound de
func VaultMirrorNotBound(mirror VaultMirror, reason, message string) VaultMirror {
	setResourceCondition(&mirror, BoundCondition, metav1.ConditionFalse, reason, message)
	return mirror
}

// VaultMirrorBound de
func VaultMirrorBound(mirror VaultMirror, reason, message string) VaultMirror {
	setResourceCondition(&mirror, BoundCondition, metav1.ConditionTrue, reason, message)
	return mirror
}

// GetStatusConditions returns a pointer to the Status.Conditions slice
func (in *VaultMirror) GetStatusConditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

type VaultMirrorVaultStatus struct {
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

// VaultMirror is the Schema for the vaultmirrors API
type VaultMirror struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VaultMirrorSpec   `json:"spec,omitempty"`
	Status VaultMirrorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VaultMirrorList contains a list of VaultMirror
type VaultMirrorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultMirror `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VaultMirror{}, &VaultMirrorList{})
}
