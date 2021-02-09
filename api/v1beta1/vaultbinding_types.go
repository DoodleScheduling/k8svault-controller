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
	// The http URL for the vault server
	// By default the global VAULT_ADDRESS gets used.
	// +optional
	Address string `json:"address,omitempty"`

	// The vault path, for example: /secret/myapp
	// +required
	Path string `json:"path"`

	// By default existing matching secrets in vault do not get overwritten
	// +optional
	ForceApply bool `json:"forceApply"`

	// Define the secrets which must be mapped to vault
	// +optional
	Fields []FieldMapping `json:"fields"`

	// The kubernetes secret the VaultBinding is referring to
	// +required
	Secret *corev1.SecretReference `json:"secret"`

	// Vault TLS configuration
	// +optional
	TLSConfig VaultTLSSpec `json:"tlsConfig"`

	// Vault authentication parameters
	// +optional
	Auth VaultAuthSpec `json:"auth,omitempty"`
}

// FieldMapping maps a secret field to the vault path
type FieldMapping struct {
	// Name is the kubernetes secret field name
	// +required
	Name string `json:"name"`

	// Rename is no required. Hovever it may be used to rewrite the field name
	// +optional
	Rename string `json:"rename,omitempty"`
}

// VaultAuthSpec is the confuguration for vault authentication which by default
// is kubernetes auth (And the only supported one in the current state)
type VaultAuthSpec struct {
	// Type is by default kubernetes authentication. The vault needs to be equipped with
	// the kubernetes auth method. Currently only kubernetes is supported.
	// +optional
	Type string `json:"type,omitempty"`

	// TokenPath allows to use a different token path used for kubernetes authentication.
	// +optional
	TokenPath string `json:"tokenPath,omitempty"`

	// Role is used to map the kubernetes serviceAccount to a vault role.
	// A default VAULT_ROLE might be set for the controller. If neither is set
	// the VaultBinding can not authenticate.
	// +optional
	Role string `json:"role,omitempty"`
}

// VaultTLSSpec Vault TLS options
type VaultTLSSpec struct {
	// +optional
	CACert string `json:"caCert,omitempty"`

	// +optional
	CAPath string `json:"caPath,omitempty"`

	// +optional
	ClientCert string `json:"clientCert,omitempty"`

	// +optional
	ClientKey string `json:"clientKey,omitempty"`

	// +optional
	ServerName string `json:"serverName,omitempty"`

	// +optional
	Insecure bool `json:"insecure,omitempty"`
}

// VaultBindingStatus defines the observed state of VaultBinding
type VaultBindingStatus struct {
	// Failures is the number of failures occured while reconciling
	Failures int64 `json:"failures,omitempty"`

	// Conditions holds the conditions for the VaultBinding.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Vault Status (not implemented yet)
	Vault VaultBindingVaultStatus `json:",inline"`
}

// Status conditions
const (
	BoundCondition = "Bound"
)

// Status reasons
const (
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