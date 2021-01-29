package controllers

import (
	"errors"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"

	infrav1beta1 "github.com/DoodleScheduling/k8svault-controller/pkg/apis/infra.doodle.com/v1beta1"
)

// Common errors
var (
	ErrInvalidFieldMapping = errors.New("Invalid field mapping provided")
	ErrNoVaultMapping      = errors.New("No vault mapping available")
)

// Mapping represents how a secret is mapped to vault. It is a representation
// of all supported annotations.
type Mapping struct {
	Vault     string
	Path      string
	Role      string
	TokenPath string
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

// Create a new mapping from a provided k8s Secret
func mapFromSecret(secret *corev1.Secret) (*Mapping, error) {
	m := NewMapping()
	if err := mapFields(m, secret); err != nil {
		return m, err
	}

	mapVault(m, secret)
	mapPath(m, secret)
	mapForce(m, secret)
	mapTLSConfig(m, secret)

	return m, nil
}

func mapFields(m *Mapping, secret *corev1.Secret) error {
	if v, ok := secret.Annotations[infrav1beta1.AnnotationFields]; ok {
		fields := strings.Split(v, ",")
		for _, v := range fields {
			pair := strings.Split(v, "=")
			switch len(pair) {
			case 1:
				m.Fields[pair[0]] = pair[0]
			case 2:
				m.Fields[pair[0]] = pair[1]
			default:
				return ErrInvalidFieldMapping
			}
		}

		return nil
	}

	// No specific fields filtered, map all secret Fields
	for k := range secret.Data {
		m.Fields[k] = k
	}

	return nil
}

func mapVault(m *Mapping, secret *corev1.Secret) {
	if v, ok := secret.Annotations[infrav1beta1.AnnotationVault]; ok {
		m.Vault = v
	}
}

func mapPath(m *Mapping, secret *corev1.Secret) {
	if v, ok := secret.Annotations[infrav1beta1.AnnotationPath]; ok {
		m.Path = v
	}
}

func mapForce(m *Mapping, secret *corev1.Secret) {
	if v, ok := secret.Annotations[infrav1beta1.AnnotationForce]; ok {
		v = strings.ToLower(v)
		if v == "1" || v == "true" || v == "yes" {
			m.Force = true
		}
	}
}

func mapTLSConfig(m *Mapping, secret *corev1.Secret) {
	//c := tlsFromViper()
	m.TLSConfig = &vaultapi.TLSConfig{}

	if v, ok := secret.Annotations[infrav1beta1.AnnotationTLSCACert]; ok {
		m.TLSConfig.CACert = v
	}
	if v, ok := secret.Annotations[infrav1beta1.AnnotationTLSCAPath]; ok {
		m.TLSConfig.CAPath = v
	}
	if v, ok := secret.Annotations[infrav1beta1.AnnotationTLSClientCert]; ok {
		m.TLSConfig.ClientCert = v
	}
	if v, ok := secret.Annotations[infrav1beta1.AnnotationTLSClientKey]; ok {
		m.TLSConfig.ClientKey = v
	}
	if v, ok := secret.Annotations[infrav1beta1.AnnotationTLSServerName]; ok {
		m.TLSConfig.TLSServerName = v
	}
	if v, ok := secret.Annotations[infrav1beta1.AnnotationTLSInsecure]; ok {
		if v == "1" || v == "true" || v == "yes" {
			m.TLSConfig.Insecure = true
		}
	}
}
