package vault

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	v1beta1 "github.com/DoodleScheduling/k8svault-controller/api/v1beta1"
)

func init() {
	registry.MustRegister("", authKubernetes)
	registry.MustRegister("kubernetes", authKubernetes)
}

const (
	serviceAccountFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	DefaultAuthRole    = "k8svault-controller"
)

type kubernetesMethod struct {
	mountPath string

	role string

	// tokenPath is an optional path to a projected service account token inside
	// the pod, for use instead of the default service account token.
	tokenPath string

	// jwtData is a ReadCloser used to inject a ReadCloser for mocking tests.
	jwtData io.ReadCloser
}

// Wrapper around vault kubernetes auth (taken from vault agent)
func authKubernetes(config *v1beta1.VaultAuthSpec) (AuthMethod, error) {
	var role string

	switch {
	case config.Role != "":
		role = config.Role
	case os.Getenv("VAULT_ROLE") != "":
		role = os.Getenv("VAULT_ROLE")
	default:
		role = DefaultAuthRole
	}

	tokenPath := config.TokenPath
	if tokenPath == "" {
		tokenPath = os.Getenv("VAULT_TOKEN_PATH")
	}

	return NewKubernetesAuthMethod(&AuthConfig{
		MountPath: "/auth/kubernetes",
		Config: map[string]interface{}{
			"role":       role,
			"token_path": tokenPath,
		},
	})
}

// NewKubernetesAuthMethod reads the user configuration and returns a configured
// AuthMethod
func NewKubernetesAuthMethod(conf *AuthConfig) (AuthMethod, error) {
	if conf == nil {
		return nil, errors.New("empty config")
	}
	if conf.Config == nil {
		return nil, errors.New("empty config data")
	}

	k := &kubernetesMethod{
		mountPath: conf.MountPath,
	}

	roleRaw, ok := conf.Config["role"]
	if !ok {
		return nil, errors.New("missing 'role' value")
	}
	k.role, ok = roleRaw.(string)
	if !ok {
		return nil, errors.New("could not convert 'role' config value to string")
	}

	tokenPathRaw, ok := conf.Config["token_path"]
	if ok {
		k.tokenPath, ok = tokenPathRaw.(string)
		if !ok {
			return nil, errors.New("could not convert 'token_path' config value to string")
		}
	}

	if k.role == "" {
		return nil, errors.New("'role' value is empty")
	}

	return k, nil
}

func (k *kubernetesMethod) Authenticate(ctx context.Context) (string, http.Header, map[string]interface{}, error) {
	jwtString, err := k.readJWT()
	if err != nil {
		return "", nil, nil, fmt.Errorf("error reading JWT with Kubernetes Auth: %q", err)
	}

	return fmt.Sprintf("%s/login", k.mountPath), nil, map[string]interface{}{
		"role": k.role,
		"jwt":  jwtString,
	}, nil
}

// readJWT reads the JWT data for the Agent to submit to  The default is
// to read the JWT from the default service account location, defined by the
// constant serviceAccountFile. In normal use k.jwtData is nil at invocation and
// the method falls back to reading the token path with os.Open, opening a file
// from either the default location or from the token_path path specified in
// configuration.
func (k *kubernetesMethod) readJWT() (string, error) {
	// load configured token path if set, default to serviceAccountFile
	tokenFilePath := serviceAccountFile
	if k.tokenPath != "" {
		tokenFilePath = k.tokenPath
	}

	data := k.jwtData
	// k.jwtData should only be non-nil in tests
	if data == nil {
		f, err := os.Open(tokenFilePath)
		if err != nil {
			return "", err
		}
		data = f
	}
	defer data.Close()

	contentBytes, err := io.ReadAll(data)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(contentBytes)), nil
}
