package vault

import (
	"fmt"
	"sync"

	v1beta1 "github.com/DoodleScheduling/k8svault-controller/api/v1beta1"
)

var registry *AuthMethodRegistry = &AuthMethodRegistry{
	methods: make(map[string]NewAuthMethod),
}

type NewAuthMethod func(conf *v1beta1.VaultAuthSpec) (AuthMethod, error)

type AuthMethodRegistry struct {
	methods map[string]NewAuthMethod
	mu      sync.Mutex
}

func (r *AuthMethodRegistry) Register(name string, init NewAuthMethod) error {
	for k := range r.methods {
		if k == name {
			return fmt.Errorf("auth method %s is already registered", name)
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.methods[name] = init
	return nil
}

func (r *AuthMethodRegistry) MustRegister(name string, init NewAuthMethod) {
	err := r.Register(name, init)
	if err != nil {
		panic(err)
	}
}

func (r *AuthMethodRegistry) Invoke(name string, conf *v1beta1.VaultAuthSpec) (AuthMethod, error) {
	for k, v := range r.methods {
		if k == name {
			return v(conf)
		}
	}

	return nil, fmt.Errorf("auth method %s is unknown", name)
}
