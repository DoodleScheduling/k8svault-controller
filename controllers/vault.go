package controllers

import (
	"context"
	"errors"
	"os"

	"github.com/go-logr/logr"
	vaultapi "github.com/hashicorp/vault/api"

	v1beta1 "github.com/DoodleScheduling/k8svault-controller/api/v1beta1"
	"github.com/DoodleScheduling/k8svault-controller/controllers/vault"
	"github.com/DoodleScheduling/k8svault-controller/controllers/vault/kubernetes"
)

const (
	// DefaultAuthRole is the vault auth role name
	DefaultAuthRole = "k8svault-controller"
)

// Common errors
var (
	ErrVaultAddrNotFound   = errors.New("Neither vault address nor a default vault address found")
	ErrFieldNotAvailable   = errors.New("Source field to be mapped does not exist")
	ErrUnsupportedAuthType = errors.New("Unsupported vault authentication")
	ErrVaultConfig         = errors.New("Failed to setup default vault configuration")
)

// Setup vault client & authentication from binding
func setupAuth(h *VaultHandler, config *v1beta1.VaultSpec) error {
	auth := vault.NewAuthHandler(&vault.AuthHandlerConfig{
		Logger: h.logger,
		Client: h.c,
	})

	var method vault.AuthMethod
	if config.Auth.Type == "kubernetes" || config.Auth.Type == "" {
		m, err := authKubernetes(h, config)
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
func authKubernetes(h *VaultHandler, config *v1beta1.VaultSpec) (vault.AuthMethod, error) {
	var role string

	switch {
	case config.Auth.Role != "":
		role = config.Auth.Role
	case os.Getenv("VAULT_ROLE") != "":
		role = os.Getenv("VAULT_ROLE")
	default:
		role = DefaultAuthRole
	}

	tokenPath := config.Auth.TokenPath
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

// NewHandler creates a vault client handler
// If the config holds no vault address it will fallback to the env VAULT_ADDRESS
func NewHandler(config *v1beta1.VaultSpec, logger logr.Logger) (*VaultHandler, error) {
	cfg := vaultapi.DefaultConfig()

	if cfg == nil {
		return nil, ErrVaultConfig
	}

	if config.Address != "" {
		cfg.Address = config.Address
	}

	// Overwrite TLS setttings with individual settings
	cfg.ConfigureTLS(convertTLSSpec(config.TLSConfig))

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

	if err = setupAuth(h, config); err != nil {
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

// VaultWriter is used to specify the destination vault
type VaultWriter interface {
	IsForceApply() bool
	GetPath() string
	GetFieldMapping() []v1beta1.FieldMapping
}

// ApplySecret applies the desired secret to vault
func (h *VaultHandler) Write(writer VaultWriter, data map[string]interface{}) (bool, error) {
	var writeBack bool

	// TODO Is there such a thing as locking the path so we don't overwrite fields which would be changed at the same time?
	dstData, err := h.Read(writer.GetPath())
	if err != nil {
		return writeBack, err
	}

	// Loop through all mapping field and apply to the vault path data
	for _, field := range writer.GetFieldMapping() {
		srcField := field.Name
		dstField := srcField
		if field.Rename != "" {
			dstField = field.Rename
		}

		h.logger.Info("applying fields to vault", "srcField", srcField, "dstField", dstField, "dstPath", writer.GetPath())

		// If k8s secret field does not exists return an error
		srcValue, ok := data[srcField]
		if !ok {
			return writeBack, ErrFieldNotAvailable
		}

		_, existingField := dstData[dstField]

		switch {
		case !existingField:
			h.logger.Info("found new field to write", "dstField", dstField)
			data[dstField] = srcValue
			writeBack = true
		case data[dstField] == srcValue:
			h.logger.Info("skipping field, no update required", "dstField", dstField)
		case writer.IsForceApply():
			data[dstField] = srcValue
			writeBack = true
		default:
			h.logger.Info("skipping field, it already exists in vault and force apply is not enabled", "dstField", dstField)
		}
	}

	if writeBack == true {
		// Finally write the secret back
		_, err = h.c.Logical().Write(writer.GetPath(), data)
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
