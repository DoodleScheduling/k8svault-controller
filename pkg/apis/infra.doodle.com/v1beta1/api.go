package v1beta1

import (
	vaultapi "github.com/hashicorp/vault/api"
)

// Supported annotations by k8svault-controller
const (
	AnnotationVault         = "k8svault-controller.v1beta1.infra.doodle.com/vault"
	AnnotationPath          = "k8svault-controller.v1beta1.infra.doodle.com/path"
	AnnotationForce         = "k8svault-controller.v1beta1.infra.doodle.com/force"
	AnnotationFields        = "k8svault-controller.v1beta1.infra.doodle.com/fields"
	AnnotationReconciledAt  = "k8svault-controller.v1beta1.infra.doodle.com/reconciledAt"
	AnnotationTLSCACert     = "k8svault-controller.v1beta1.infra.doodle.com/tlsCACert"
	AnnotationTLSCAPath     = "k8svault-controller.v1beta1.infra.doodle.com/tlsCAPath"
	AnnotationTLSClientCert = "k8svault-controller.v1beta1.infra.doodle.com/tlsClientCert"
	AnnotationTLSClientKey  = "k8svault-controller.v1beta1.infra.doodle.com/tlsClientKey"
	AnnotationTLSServerName = "k8svault-controller.v1beta1.infra.doodle.com/tlsServerName"
	AnnotationTLSInsecure   = "k8svault-controller.v1beta1.infra.doodle.com/tlsInsecure"
)

// Mapping represents how a secret is mapped to vault. It is a representation
// of all supported annotations.
type Mapping struct {
	Vault     string
	Path      string
	Force     bool
	TLSConfig *vaultapi.TLSConfig
	Fields    map[string]string
}

// NewMapping creates a new mapping
func NewMapping() *Mapping {
	return &Mapping{
		Fields: make(map[string]string),
	}
}
