package controllers

import (
	"encoding/base64"
	"errors"

	"github.com/go-logr/logr"
	"github.com/hashicorp/vault/api"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"

	infrav1beta1 "github.com/DoodleScheduling/k8svault-controller/pkg/apis/infra.doodle.com/v1beta1"
)

// Common errors
var (
	ErrVaultAddrNotFound          = errors.New("Neither vault annotation nor a default vault address found")
	ErrK8sSecretFieldNotAvailable = errors.New("K8s secret field to be mapped does not exist")
)

// Vault handler
type Vault struct {
	c *api.Client
	m *infrav1beta1.Mapping
	l logr.Logger
}

func tlsFromViper() *vaultapi.TLSConfig {
	return &vaultapi.TLSConfig{
		CACert:        viper.GetString("tls-cacert"),
		CAPath:        viper.GetString("tls-capath"),
		ClientCert:    viper.GetString("tls-client-cert"),
		ClientKey:     viper.GetString("tls-client-key"),
		TLSServerName: viper.GetString("tls-server-name"),
		Insecure:      viper.GetBool("tls-insecure"),
	}
}

// FromMapping creates a vault client from Kubernetes to Vault mapping
// If the mapping holds no vault address it will fallback to the env VAULT_ADDRESS
func FromMapping(m *infrav1beta1.Mapping) (*Vault, error) {
	c := vaultapi.DefaultConfig()

	switch {
	case m.Vault != "":
		c.Address = m.Vault
	case viper.GetString("vault-addr") != "":
		c.Address = viper.GetString("vault-addr")
	default:
		return nil, ErrVaultAddrNotFound
	}

	c.ConfigureTLS(m.TLSConfig)
	client, err := api.NewClient(c)
	if err != nil {
		return nil, err
	}

	return &Vault{
		c: client,
		m: m,
		l: &logr.DiscardLogger{},
	}, nil
}

// WithLogger inject logr compatible logger
func (v *Vault) WithLogger(l logr.Logger) *Vault {
	v.l = l
	return v
}

// ApplySecret applies the desired secret to vault
func (v *Vault) ApplySecret(secret *corev1.Secret) error {
	data, err := v.Read(v.m.Path)
	if err != nil {
		return err
	}

	// Loop through all mapping field and apply to the vault path data
	for k8sField, vaultField := range v.m.Fields {
		v.l.Info("Applying k8s field to vault", "k8sField", k8sField, "vaultField", vaultField, "vaultPath", v.m.Path)

		// If k8s secret field does not exists return an error
		k8sValue, ok := secret.Data[k8sField]
		if !ok {
			return ErrK8sSecretFieldNotAvailable
		}

		secret, err := base64.StdEncoding.DecodeString(string(k8sValue))
		if err != nil {
			return err
		}

		//  Don't overwrite vault field if force is false
		if _, ok := data[vaultField]; ok {
			if v.m.Force == true {
				data[vaultField] = secret
			} else {
				v.l.Info("Skipping field, it already exists in vault and force apply is disabled", "vaultField", vaultField)
			}
		} else {
			data[vaultField] = secret
		}
	}

	// Finally write the secret back
	_, err = v.c.Logical().Write(v.m.Path, data)
	return err
}

// Read vault path and return data map
// Return empty map if no data exists
func (v *Vault) Read(path string) (map[string]interface{}, error) {
	s, err := v.c.Logical().Read(path)
	if err != nil {
		return nil, err
	}

	// Return empty map if no data exists
	if s == nil || s.Data == nil {
		return make(map[string]interface{}), nil
	}

	return s.Data, nil
}
