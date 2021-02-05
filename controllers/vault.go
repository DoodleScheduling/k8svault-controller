package controllers

import (
	"context"
	"errors"
	"os"

	"github.com/go-logr/logr"
	vaultapi "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"

	v1beta1 "github.com/DoodleScheduling/k8svault-controller/api/v1beta1"
	"github.com/DoodleScheduling/k8svault-controller/controllers/vault"
	"github.com/DoodleScheduling/k8svault-controller/controllers/vault/kubernetes"
)

// Common errors
var (
	ErrVaultAddrNotFound          = errors.New("Neither vault address nor a default vault address found")
	ErrK8sSecretFieldNotAvailable = errors.New("K8s secret field to be mapped does not exist")
	ErrUnsupportedAuthType        = errors.New("Unsupported vault authentication")
	ErrVaultConfig                = errors.New("Failed to setup default vault configuration")
)

// Setup vault client & authentication from binding
func setupAuth(h *VaultHandler, binding *v1beta1.VaultBinding) error {
	auth := vault.NewAuthHandler(&vault.AuthHandlerConfig{
		Logger: h.logger,
		Client: h.c,
	})

	var method vault.AuthMethod

	if binding.Spec.Auth.Type == "kubernetes" || binding.Spec.Auth.Type == "" {
		m, err := authKubernetes(h, binding)
		if err != nil {
			return err
		}

		method = m
	} else {
		return ErrUnsupportedAuthType
	}

	if err := auth.Authenticate(context.TODO(), method); err != nil {
		return err
	}

	return nil
}

// Wrapper around vault kubernetes auth (taken from vault agent)
// Injects env variables if not set on the binding
func authKubernetes(h *VaultHandler, binding *v1beta1.VaultBinding) (vault.AuthMethod, error) {
	role := binding.Spec.Auth.Role
	if role == "" {
		role = os.Getenv("VAULT_ROLE")
	}

	tokenPath := binding.Spec.Auth.TokenPath
	if tokenPath == "" {
		tokenPath = os.Getenv("VAULT_TOKEN_PATH")
	}

	return kubernetes.NewKubernetesAuthMethod(&vault.AuthConfig{
		Logger:    h.logger,
		MountPath: "/auth/kubernetes",
		Config: map[string]interface{}{
			"role":       role,
			"token_path": tokenPath,
		},
	})
}

func convertTLSSpec(spec v1beta1.VaultTLSSpec) *vaultapi.TLSConfig {
	return &vaultapi.TLSConfig{
		CACert:        spec.CACert,
		ClientCert:    spec.ClientCert,
		ClientKey:     spec.ClientKey,
		TLSServerName: spec.ServerName,
		Insecure:      spec.Insecure,
	}
}

// FromBinding creates a vault client handler
// If the binding holds no vault address it will fallback to the env VAULT_ADDRESS
func FromBinding(binding *v1beta1.VaultBinding, logger logr.Logger) (*VaultHandler, error) {
	cfg := vaultapi.DefaultConfig()

	if cfg == nil {
		return nil, ErrVaultConfig
	}

	if binding.Spec.Address != "" {
		cfg.Address = binding.Spec.Address
	}

	// Overwrite TLS setttings with individual settings
	cfg.ConfigureTLS(convertTLSSpec(binding.Spec.TLSConfig))

	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	h := &VaultHandler{
		cfg:    cfg,
		c:      client,
		logger: logger,
	}

	logger.Info("setup vault client", "vault", cfg.Address)

	if err = setupAuth(h, binding); err != nil {
		return nil, err
	}

	return h, nil
}

// VaultHandler
type VaultHandler struct {
	c      *vaultapi.Client
	cfg    *vaultapi.Config
	auth   *vault.AuthHandler
	logger logr.Logger
}

// ApplySecret applies the desired secret to vault
func (h *VaultHandler) ApplySecret(binding *v1beta1.VaultBinding, secret *corev1.Secret) (bool, error) {
	var writeBack bool

	// TODO Is there such a thing as locking the path so we don't overwrite fields which would be changed at the same time?
	data, err := h.Read(binding.Spec.Path)
	if err != nil {
		return writeBack, err
	}

	// Loop through all mapping field and apply to the vault path data
	for _, field := range binding.Spec.Fields {
		k8sField := field.Name
		vaultField := k8sField
		if field.Rename != "" {
			vaultField = field.Rename
		}

		h.logger.Info("applying k8s field to vault", "k8sField", k8sField, "vaultField", vaultField, "vaultPath", binding.Spec.Path)

		// If k8s secret field does not exists return an error
		k8sValue, ok := secret.Data[k8sField]
		if !ok {
			return writeBack, ErrK8sSecretFieldNotAvailable
		}

		secret := string(k8sValue)

		_, existingField := data[vaultField]

		switch {
		case !existingField:
			h.logger.Info("found new field to write", "vaultField", vaultField)
			data[vaultField] = secret
			writeBack = true
		case data[vaultField] == secret:
			h.logger.Info("skipping field, no update required", "vaultField", vaultField)
		case binding.Spec.ForceApply == true:
			data[vaultField] = secret
			writeBack = true
		default:
			h.logger.Info("skipping field, it already exists in vault and force apply is disabled", "vaultField", vaultField)
		}
	}

	if writeBack == true {
		// Finally write the secret back
		_, err = h.c.Logical().Write(binding.Spec.Path, data)
		if err != nil {
			return writeBack, err
		}
	}

	return writeBack, nil
}

// Read vault path and return data map
// Return empty map if no data exists
func (h *VaultHandler) Read(path string) (map[string]interface{}, error) {
	s, err := h.c.Logical().Read(path)
	if err != nil {
		return nil, err
	}

	// Return empty map if no data exists
	if s == nil || s.Data == nil {
		return make(map[string]interface{}), nil
	}

	return s.Data, nil
}
