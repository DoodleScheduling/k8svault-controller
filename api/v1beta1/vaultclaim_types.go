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

// VaultClaimSpec defines the desired state of VaultClaim
type VaultClaimSpec struct {
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

// VaultClaimStatus defines the observed state of VaultClaim
type VaultClaimStatus struct {
	// ObservedGeneration is the last observed generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions holds the conditions for the HelmRelease.
	// +optional
	//Conditions []metav1.Condition `json:"conditions,omitempty"`

	// LastAppliedRevision is the revision of the last successfully applied source.
	// +optional
	LastAppliedRevision string `json:"lastAppliedRevision,omitempty"`

	// LastAttemptedRevision is the revision of the last reconciliation attempt.
	// +optional
	LastAttemptedRevision string `json:"lastAttemptedRevision,omitempty"`

	Vault VaultClaimVaultStatus `json:",inline"`
}

const (
	BoundCondition              = "Bound"
	VaultConnectionFailedReason = "VaultConnectionFailed"
	VaultUpdateFailedReason     = "VaultUpdateFailed"
	VaultUpdateSuccessfulReason = "VaultUpdateSuccessful"
	SecretNotFoundReason        = "SecretNotFoundFailed"
)

// ObjectWithStatusConditions is an interface that describes kubernetes resource
// type structs with Status Conditions
type ObjectWithStatusConditions interface {
	GetStatusConditions() *[]metav1.Condition
}

// setResourceCondition sets the given condition with the given status,
// reason and message on a resource.
func setResourceCondition(obj ObjectWithStatusConditions, condition string, status metav1.ConditionStatus, reason, message string) {
	conditions := obj.GetStatusConditions()

	newCondition := metav1.Condition{
		Type:    condition,
		Status:  status,
		Reason:  reason,
		Message: message,
	}

	apimeta.SetStatusCondition(conditions, newCondition)
}

func VaultClaimNotBound(claim *VaultClaim, reason, message string) HelmRelease {
	setResourceCondition(claim, BoundCondition, metav1.ConditionFalse, reason, message)
	claim.Status.Failures++
	return claim
}

func VaultClaimBound(claim *VaultClaim, reason, message string) HelmRelease {
	setResourceCondition(claim, BoundCondition, metav1.ConditionTrue, reason, message)
	claim.Status.Failures++
	return claim
}

type VaultClaimVaultStatus struct {
	Address    string `json:"address,omitempty"`
	Path       string `json:"path,omitempty"`
	ForceApply bool   `json:"forceApply,omitempty"`
	Fields     string `json:"fields,omitempty"`
}

// +kubebuilder:object:root=true

// VaultClaim is the Schema for the vaultclaims API
type VaultClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VaultClaimSpec   `json:"spec,omitempty"`
	Status VaultClaimStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VaultClaimList contains a list of VaultClaim
type VaultClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultClaim `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VaultClaim{}, &VaultClaimList{})
}
