package vault

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"github.com/hashicorp/vault/api"
	vaultapi "github.com/hashicorp/vault/api"

	v1beta1 "github.com/DoodleScheduling/k8svault-controller/api/v1beta1"
)

// Common errors
var (
	ErrVaultAddrNotFound   = errors.New("Neither vault address nor a default vault address found")
	ErrFieldNotAvailable   = errors.New("Source field to be mapped does not exist")
	ErrUnsupportedAuthType = errors.New("Unsupported vault authentication")
	ErrVaultConfig         = errors.New("Failed to setup default vault configuration")
	ErrPathNotFound        = errors.New("Vault path not found")
)

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
		c:      client.Logical(),
		logger: logger,
	}

	logger.Info("setup vault client", "vault", cfg.Address)

	opts := AuthHandlerConfig{
		Writer:      client.Logical(),
		TokenWriter: client,
	}

	if err = setupAuth(h, opts, &config.Auth); err != nil {
		return nil, err
	}

	return h, nil
}

type Writer interface {
	Write(path string, data map[string]interface{}) (*api.Secret, error)
}

type Reader interface {
	Read(path string) (*api.Secret, error)
}

type ReadWriter interface {
	Reader
	Writer
}

// Mapper retrieves mapping configuration
type Mapper interface {
	IsForceApply() bool
	GetPath() string
	GetFieldMapping() []v1beta1.FieldMapping
}

// VaultHandler
type VaultHandler struct {
	c      ReadWriter
	cfg    *vaultapi.Config
	auth   *AuthHandler
	logger logr.Logger
}

// Write writes secrets to vault defined by the mapper
func (h *VaultHandler) Write(writer Mapper, srcData map[string]interface{}) (bool, error) {
	var writeBack bool

	// Ignore error if there is no path at the destination
	data, err := h.Read(writer.GetPath())
	if err != nil && err != ErrPathNotFound {
		return writeBack, err
	}

	// If no field mapping is configured all fields get mapped with their source field name
	mapping := writer.GetFieldMapping()
	if len(mapping) == 0 {
		for k, _ := range srcData {
			mapping = append(mapping, v1beta1.FieldMapping{
				Name: k,
			})
		}
	}

	// Loop through all mapping field and apply to the vault path data
	for _, field := range mapping {
		srcField := field.Name
		dstField := srcField
		if field.Rename != "" {
			dstField = field.Rename
		}

		h.logger.Info("applying fields to vault", "srcField", srcField, "dstField", dstField, "dstPath", writer.GetPath())

		// If k8s secret field does not exists return an error
		srcValue, ok := srcData[srcField]
		if !ok {
			return writeBack, ErrFieldNotAvailable
		}

		_, existingField := data[dstField]

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
		_, err = h.c.Write(writer.GetPath(), data)
		if err != nil {
			return writeBack, err
		}
	}

	return writeBack, nil
}

// Read vault path and return data map
// Return empty map if no data exists
func (h *VaultHandler) Read(path string) (map[string]interface{}, error) {
	s, err := h.c.Read(path)
	if err != nil {
		return nil, err
	}

	// Return empty map and PathNotFound error
	if s == nil || s.Data == nil {
		return make(map[string]interface{}), ErrPathNotFound
	}

	return s.Data, nil
}

// Setup vault client & authentication from binding
func setupAuth(h *VaultHandler, opts AuthHandlerConfig, config *v1beta1.VaultAuthSpec) error {
	handler := NewAuthHandler(opts)
	method, err := registry.Invoke(config.Type, config)

	if err != nil {
		return err
	}

	if err := handler.Authenticate(context.TODO(), method); err != nil {
		return err
	}

	return nil
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
