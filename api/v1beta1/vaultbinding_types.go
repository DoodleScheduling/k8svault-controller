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
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VaultBindingSpec defines the desired state of VaultBinding
type VaultBindingSpec struct {
	Address    string                  `json:"address,omitempty"`
	Path       string                  `json:"path,omitempty"`
	ForceApply bool                    `json:"forceApply,omitempty"`
	Fields     []FieldMapping          `json:"fields,omitempty"`
	Secret     *corev1.SecretReference `json:"secret,omitempty"`
	TLSConfig  VaultTLSSpec            `json:"tlsConfig,omitempty"`
	Auth       VaultAuthSpec           `json:"auth,omitempty"`
}

type FieldMapping struct {
	Name   string `json:"name,omitempty"`
	Rename string `json:"rename,omitempty"`
}

type VaultAuthSpec struct {
	Type      string `json:"type,omitempty"`
	TokenPath string `json:"tokenPath,omitempty"`
	Role      string `json:"role,omitempty"`
}

type VaultTLSSpec struct {
	CACert     string `json:"caCert,omitempty"`
	CAPath     string `json:"caPath,omitempty"`
	ClientCert string `json:"clientCert,omitempty"`
	ClientKey  string `json:"clientKey,omitempty"`
	ServerName string `json:"serverName,omitempty"`
	Insecure   bool   `json:"insecure,omitempty"`
}

// VaultBindingStatus defines the observed state of VaultBinding
type VaultBindingStatus struct {
	// Failures is the number of failures occured while reconciling
	Failures int64 `json:"failures,omitempty"`

	// ObservedGeneration is the last observed generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions holds the conditions for the VaultBinding.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// LastAppliedRevision is the revision of the last successfully applied source.
	// +optional
	LastAppliedRevision string `json:"lastAppliedRevision,omitempty"`

	// LastAttemptedRevision is the revision of the last reconciliation attempt.
	// +optional
	LastAttemptedRevision string `json:"lastAttemptedRevision,omitempty"`

	Vault VaultBindingVaultStatus `json:",inline"`
}

const (
	BoundCondition              = "Bound"
	VaultConnectionFailedReason = "VaultConnectionFailed"
	VaultUpdateFailedReason     = "VaultUpdateFailed"
	VaultUpdateSuccessfulReason = "VaultUpdateSuccessful"
	SecretNotFoundReason        = "SecretNotFoundFailed"
)

// setResourceCondition sets the given condition with the given status,
// reason and message on a resource.
func setResourceCondition(binding *VaultBinding, condition string, status metav1.ConditionStatus, reason, message string) {
	conditions := binding.GetStatusConditions()

	newCondition := metav1.Condition{
		Type:    condition,
		Status:  status,
		Reason:  reason,
		Message: message,
	}

	apimeta.SetStatusCondition(conditions, newCondition)
}

// VaultBindingNotBound de
func VaultBindingNotBound(binding *VaultBinding, reason, message string) *VaultBinding {
	setResourceCondition(binding, BoundCondition, metav1.ConditionFalse, reason, message)
	binding.Status.Failures++
	return binding
}

// VaultBindingBound de
func VaultBindingBound(binding *VaultBinding, reason, message string) *VaultBinding {
	setResourceCondition(binding, BoundCondition, metav1.ConditionTrue, reason, message)
	binding.Status.Failures = 0
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
