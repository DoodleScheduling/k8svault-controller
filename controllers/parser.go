package controllers

import (
	"strings"

	corev1 "k8s.io/api/core/v1"

	infrav1beta1 "github.com/DoodleScheduling/k8svault-controller/pkg/apis/infra.doodle.com/v1beta1"
)

// Create a new mapping from a provided k8s Secret
func mapFromSecret(secret *corev1.Secret) (*infrav1beta1.Mapping, error) {
	m := infrav1beta1.NewMapping()
	if err := mapFields(m, secret); err != nil {
		return m, err
	}

	mapVault(m, secret)
	mapPath(m, secret)
	mapForce(m, secret)

	return m, nil
}

func mapFields(m *infrav1beta1.Mapping, secret *corev1.Secret) error {
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

func mapVault(m *infrav1beta1.Mapping, secret *corev1.Secret) {
	if v, ok := secret.Annotations[infrav1beta1.AnnotationVault]; ok {
		m.Vault = v
	}
}

func mapPath(m *infrav1beta1.Mapping, secret *corev1.Secret) {
	if v, ok := secret.Annotations[infrav1beta1.AnnotationPath]; ok {
		m.Path = v
	}
}

func mapForce(m *infrav1beta1.Mapping, secret *corev1.Secret) {
	if v, ok := secret.Annotations[infrav1beta1.AnnotationForce]; ok {
		v = strings.ToLower(v)
		if v == "1" || v == "true" || v == "yes" {
			m.Force = true
		}
	}
}
