package controllers

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	vaultapi "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/DoodleScheduling/k8svault-controller/controllers/vault"
	"github.com/DoodleScheduling/k8svault-controller/controllers/vault/kubernetes"
)

// Common errors
var (
	ErrVaultAddrNotFound          = errors.New("Neither vault annotation nor a default vault address found")
	ErrK8sSecretFieldNotAvailable = errors.New("K8s secret field to be mapped does not exist")
)

// Make sure we only have one client per setting construct
// Create vault client and start token lifecycle manager
func setupClient(cfg *vaultapi.Config, m *Mapping, logger logr.Logger) (*Vault, error) {
	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	auth := vault.NewAuthHandler(&vault.AuthHandlerConfig{
		Logger: logger,
		Client: client,
	})

	// Vault role is not a common role configured by vault api DefaultConfig()
	// It is only used certain auth mechanisms
	role := m.Role
	if role == "" {
		role = os.Getenv("VAULT_ROLE")
	}

	tokenPath := m.TokenPath
	if tokenPath == "" {
		tokenPath = os.Getenv("VAULT_TOKEN_PATH")
	}

	method, err := kubernetes.NewKubernetesAuthMethod(&vault.AuthConfig{
		Logger:    logger,
		MountPath: "/auth/kubernetes",
		Config: map[string]interface{}{
			"role":       role,
			"token_path": tokenPath,
		},
	})

	if err != nil {
		return nil, err
	}

	if err := auth.Authenticate(context.TODO(), method); err != nil {
		return nil, err
	}

	v := &Vault{
		c:      client,
		cfg:    cfg,
		m:      m,
		logger: logger,
	}

	return v, nil
}

// FromMapping creates a vault client from Kubernetes to Vault mapping
// If the mapping holds no vault address it will fallback to the env VAULT_ADDRESS
func FromMapping(m *Mapping, logger logr.Logger) (*Vault, error) {
	cfg := vaultapi.DefaultConfig()

	if m.Vault != "" {
		cfg.Address = m.Vault
	}

	// Overwrite TLS setttings with individual settings
	cfg.ConfigureTLS(m.TLSConfig)

	client, err := setupClient(cfg, m, logger)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Vault handler
type Vault struct {
	c      *vaultapi.Client
	cfg    *vaultapi.Config
	auth   *vault.AuthHandler
	m      *Mapping
	logger logr.Logger
}

// ApplySecret applies the desired secret to vault
func (v *Vault) ApplySecret(m *Mapping, secret *corev1.Secret, rec record.EventRecorder) error {
	var writeBack bool

	// TODO Is there such a thing as locking the path so we don't overwrite fields which would be changed at the same time?
	data, err := v.Read(m.Path)
	if err != nil {
		return err
	}

	// Loop through all mapping field and apply to the vault path data
	for k8sField, vaultField := range m.Fields {
		v.logger.Info("Applying k8s field to vault", "k8sField", k8sField, "vaultField", vaultField, "vaultPath", m.Path)

		// If k8s secret field does not exists return an error
		k8sValue, ok := secret.Data[k8sField]
		if !ok {
			return ErrK8sSecretFieldNotAvailable
		}
		v.logger.Info("Applying k8s field to vault", "trest", string(k8sValue), "s2", k8sValue, "s", secret.Data)

		secret := string(k8sValue)

		_, existingField := data[vaultField]

		switch {
		case !existingField:
			v.logger.Info("Found new field to write", "vaultField", vaultField)
			data[vaultField] = secret
			writeBack = true
		case data[vaultField] == secret:
			v.logger.Info("Skipping field, no update required", "vaultField", vaultField)
		case m.Force == true:
			data[vaultField] = secret
			writeBack = true
		default:
			v.logger.Info("Skipping field, it already exists in vault and force apply is disabled", "vaultField", vaultField)
		}
	}

	if writeBack == true {
		// Finally write the secret back
		_, err = v.c.Logical().Write(m.Path, data)
		if err != nil {
			return err
		}

		rec.Event(secret, "Normal", "errror", fmt.Sprintf("Synced secret %s/%s fields to vault", secret.Namespace, secret.Name))
	}

	return nil
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
