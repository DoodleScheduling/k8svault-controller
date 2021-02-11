package v1beta1

import (
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	// the VaultMirror can not authenticate.
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

// FieldMapping maps a secret field to the vault path
type FieldMapping struct {
	// Name is the kubernetes secret field name
	// +required
	Name string `json:"name"`

	// Rename is no required. Hovever it may be used to rewrite the field name
	// +optional
	Rename string `json:"rename,omitempty"`
}

// ConditionalResource is a resource with conditions
type ConditionalResource interface {
	GetStatusConditions() *[]metav1.Condition
}

// setResourceCondition sets the given condition with the given status,
// reason and message on a resource.
func setResourceCondition(resource ConditionalResource, condition string, status metav1.ConditionStatus, reason, message string) {
	conditions := resource.GetStatusConditions()

	newCondition := metav1.Condition{
		Type:    condition,
		Status:  status,
		Reason:  reason,
		Message: message,
	}

	apimeta.SetStatusCondition(conditions, newCondition)
}
