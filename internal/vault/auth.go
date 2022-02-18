// This package implements authentication for vault
// Most of this code within this package has been borrowed and shrinked from the vault client agent.
// See https://github.com/hashicorp/vault/blob/master/command/agent/auth/auth.go

// Note this package uses API which is in the vault stable release but it was not released in the api package,
// see https://github.com/hashicorp/vault/issues/10490

package vault

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// AuthMethod is the interface that auto-auth methods implement for the agent
// to use.
type AuthMethod interface {
	// Authenticate returns a mount path, header, request body, and error.
	// The header may be nil if no special header is needed.
	Authenticate(context.Context) (string, http.Header, map[string]interface{}, error)
}

type AuthConfig struct {
	MountPath string
	Config    map[string]interface{}
}

// AuthHandler is responsible for keeping a token alive and renewed and passing
// new tokens to the sink server
type AuthHandler struct {
	writer      Writer
	tokenWriter TokenWriter
}

type TokenWriter interface {
	SetToken(token string)
}

type AuthHandlerConfig struct {
	Writer      Writer
	TokenWriter TokenWriter
}

func NewAuthHandler(opts AuthHandlerConfig) *AuthHandler {
	ah := &AuthHandler{
		writer:      opts.Writer,
		tokenWriter: opts.TokenWriter,
	}

	return ah
}

func (ah *AuthHandler) Authenticate(ctx context.Context, am AuthMethod) error {
	if am == nil {
		return errors.New("auth handler: nil auth method")
	}

	path, _, data, err := am.Authenticate(ctx)

	if err != nil {
		return fmt.Errorf("error getting path or data from method: %w", err)
	}

	secret, err := ah.writer.Write(path, data)

	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return errors.New("authentication returned nil auth info")
	}

	if secret.Auth.ClientToken == "" {
		return errors.New("authentication returned empty client token")
	}

	ah.tokenWriter.SetToken(secret.Auth.ClientToken)
	return nil
}
