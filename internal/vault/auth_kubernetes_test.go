package vault

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestNewKubernetesAuthMethod(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name        string
		config      *AuthConfig
		expectError error
	}{
		{
			name:        "fails if config is nil",
			config:      nil,
			expectError: errors.New("empty config"),
		},
		{
			name:        "fails if config does not contain config key",
			config:      &AuthConfig{},
			expectError: errors.New("empty config data"),
		},
		{
			name: "fails if config is missing role",
			config: &AuthConfig{
				Config: make(map[string]interface{}),
			},
			expectError: errors.New("missing 'role' value"),
		},
		{
			name: "fails if config role is not a string",
			config: &AuthConfig{
				Config: map[string]interface{}{
					"role": 1,
				},
			},
			expectError: errors.New("could not convert 'role' config value to string"),
		},
		{
			name: "fails if config token_path is not a string",
			config: &AuthConfig{
				Config: map[string]interface{}{
					"role":       "strawberry",
					"token_path": 1,
				},
			},
			expectError: errors.New("could not convert 'token_path' config value to string"),
		},
		{
			name: "fails if role is empty string",
			config: &AuthConfig{
				Config: map[string]interface{}{
					"role":       "",
					"token_path": "path",
				},
			},
			expectError: errors.New("'role' value is empty"),
		},
		{
			name: "Initializes successfully with given config",
			config: &AuthConfig{
				Config: map[string]interface{}{
					"role":       "test",
					"token_path": "path",
				},
			},
			expectError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewKubernetesAuthMethod(test.config)
			if test.expectError == nil {
				g.Expect(err).NotTo(HaveOccurred(), "error occurd during initialize kubernetes auth but should not")
			} else {
				g.Expect(err.Error()).To(Equal(test.expectError.Error()))
			}
		})
	}
}

func TestAuthKubernetes(t *testing.T) {
	g := NewWithT(t)

	file, err := ioutil.TempFile(os.TempDir(), "jwt")
	g.Expect(err).NotTo(HaveOccurred(), "failed creating test jwt file")
	defer os.Remove(file.Name())
	file.Write([]byte("strawberry"))

	handler, err := NewKubernetesAuthMethod(&AuthConfig{
		MountPath: "/berries",
		Config: map[string]interface{}{
			"role":       "blueberry",
			"token_path": file.Name(),
		},
	})

	g.Expect(err).NotTo(HaveOccurred(), "error occurd during initialize kubernetes auth but should not")
	path, _, config, err := handler.Authenticate(context.TODO())
	g.Expect(err).NotTo(HaveOccurred(), "error reading kubernetes jwt")
	g.Expect(path).To(Equal("/berries/login"))
	g.Expect(config["jwt"]).To(Equal("strawberry"))
	g.Expect(config["role"]).To(Equal("blueberry"))
}
