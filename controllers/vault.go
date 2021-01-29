package controllers

import (
	"context"
	"errors"
	"os"
	"sync"

	"github.com/go-logr/logr"
	vaultapi "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"

	"github.com/DoodleScheduling/k8svault-controller/controllers/vault"
	"github.com/DoodleScheduling/k8svault-controller/controllers/vault/kubernetes"
)

// Common errors
var (
	ErrVaultAddrNotFound          = errors.New("Neither vault annotation nor a default vault address found")
	ErrK8sSecretFieldNotAvailable = errors.New("K8s secret field to be mapped does not exist")
)

// ClientRegistry holds a client per vault configuration
type ClientRegistry struct {
	logger   logr.Logger
	registry []*Vault
	mutex    *sync.Mutex
}

// WithLogger inject logr compatible logger
func (r *ClientRegistry) WithLogger(l logr.Logger) *ClientRegistry {
	r.logger = l
	return r
}

// NewClientRegistry returns a new vault client registry
func NewClientRegistry() *ClientRegistry {
	return &ClientRegistry{
		logger: &logr.DiscardLogger{},
		mutex:  &sync.Mutex{},
	}
}

// Make sure we only have one client per setting construct
// Create vault client and start token lifecycle manager
func (r *ClientRegistry) setupClient(cfg *vaultapi.Config, m *Mapping) (*Vault, error) {
	// Return existing client if available
	for _, v := range r.registry {
		r.logger.Info("regi", "exist_vault", v.m.Vault, "newv", m.Vault, "exuist_role", v.m.Role, "newfr", m.Role, "exixt_pat", v.m.TokenPath, "new-poa", m.TokenPath)

		if v.m.IsSame(m) {
			r.logger.Info("Secret is already using a known client")
			return v, nil
		}
	}

	r.logger.Info("Preparing new vault client")
	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	auth := vault.NewAuthHandler(&vault.AuthHandlerConfig{
		Logger: r.logger,
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
		Logger:    r.logger,
		MountPath: "/auth/kubernetes",
		Config: map[string]interface{}{
			"role":       role,
			"token_path": tokenPath,
		},
	})

	if err != nil {
		return nil, err
	}

	v := &Vault{
		c:    client,
		cfg:  cfg,
		auth: auth,
		m:    m,
		// We inject a /dev/null logger here because the registry which is also responsible to delegate token renewals
		// logs with other context (global context) while a vault mapping instance logs on a per secret context
		// You can inject a separate logger into Vault{}
		logger: &logr.DiscardLogger{},
	}

	// Start auth lifecycle
	// If any error occurs we're going to remove the client from the registry
	// and let it recreate in the next reconcilation as settings might have changed.
	go func(v *Vault, logr.Logger) {
		r.logger.Info("Starting token auth lifecycle manager")
		err := auth.Run(context.TODO(), method)
		if err != nil {
			r.logger.Error(error, "Got error from token lifecycle manager")

			r.mutex.Lock()
			for k,regClient := range r.registry {
				if v == regClient {
					r.registry = append(s[:index], s[index+1:]...)
				}
			}
			r.mutex.Unlock()
		}
	}(v, logger)

	r.mutex.Lock()
	r.registry = append(r.registry, v)
	r.mutex.Unlock()

	return v, nil
}

// FromMapping creates a vault client from Kubernetes to Vault mapping
// If the mapping holds no vault address it will fallback to the env VAULT_ADDRESS
func (r *ClientRegistry) FromMapping(m *Mapping) (*Vault, error) {
	cfg := vaultapi.DefaultConfig()

	if m.Vault != "" {
		cfg.Address = m.Vault
	}

	// Overwrite TLS setttings with individual settings
	cfg.ConfigureTLS(m.TLSConfig)

	client, err := r.setupClient(cfg, m)
	if err != nil {
		return nil, err
	}

	// Check if our token sink has a new token
	/*var token string
	token = <-client.auth.OutputCh
	r.logger.Info("TOKEN", "toklen", token)
	client.c.SetToken(token)*/

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

// WithLogger inject logr compatible logger
func (v *Vault) WithLogger(l logr.Logger) *Vault {
	v.logger = l
	return v
}

// ApplySecret applies the desired secret to vault
func (v *Vault) ApplySecret(m *Mapping, secret *corev1.Secret) error {
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

		//  Don't overwrite vault field if force is false
		if _, ok := data[vaultField]; ok {
			if m.Force == true {
				data[vaultField] = secret
			} else {
				v.logger.Info("Skipping field, it already exists in vault and force apply is disabled", "vaultField", vaultField)
			}
		} else {
			data[vaultField] = secret
		}
	}

	// Finally write the secret back
	_, err = v.c.Logical().Write(m.Path, data)
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
