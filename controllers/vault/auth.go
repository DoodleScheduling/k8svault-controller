// This package implements authentication for vault
// Most of this code within this package has been borrowed from the vault client agent.
// See https://github.com/hashicorp/vault/blob/master/command/agent/auth/auth.go

// Note this package uses API which is in the vault stable release but it was not released in the api package,
// see https://github.com/hashicorp/vault/issues/10490

package vault

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	vaultapi "github.com/hashicorp/vault/api"
)

// AuthMethod is the interface that auto-auth methods implement for the agent
// to use.
type AuthMethod interface {
	// Authenticate returns a mount path, header, request body, and error.
	// The header may be nil if no special header is needed.
	Authenticate(context.Context, *vaultapi.Client) (string, http.Header, map[string]interface{}, error)
	NewCreds() chan struct{}
	CredSuccess()
	Shutdown()
}

// AuthMethodWithClient is an extended interface that can return an API client
// for use during the authentication call.
type AuthMethodWithClient interface {
	AuthMethod
	AuthClient(client *vaultapi.Client) (*vaultapi.Client, error)
}

type AuthConfig struct {
	Logger    logr.Logger
	MountPath string
	WrapTTL   time.Duration
	Config    map[string]interface{}
}

// AuthHandler is responsible for keeping a token alive and renewed and passing
// new tokens to the sink server
type AuthHandler struct {
	//OutputCh                     chan string
	//TemplateTokenCh              chan string
	logger                       logr.Logger
	client                       *vaultapi.Client
	random                       *rand.Rand
	wrapTTL                      time.Duration
	enableReauthOnNewCredentials bool
	enableTemplateTokenCh        bool
}

type AuthHandlerConfig struct {
	Logger                       logr.Logger
	Client                       *vaultapi.Client
	WrapTTL                      time.Duration
	EnableReauthOnNewCredentials bool
	//EnableTemplateTokenCh        bool
}

func NewAuthHandler(conf *AuthHandlerConfig) *AuthHandler {
	ah := &AuthHandler{
		//TemplateTokenCh:              make(chan string, 1),
		logger:                       conf.Logger,
		client:                       conf.Client,
		random:                       rand.New(rand.NewSource(int64(time.Now().Nanosecond()))),
		wrapTTL:                      conf.WrapTTL,
		enableReauthOnNewCredentials: conf.EnableReauthOnNewCredentials,
		//enableTemplateTokenCh:        conf.EnableTemplateTokenCh,
	}

	return ah
}

func (ah *AuthHandler) Run(ctx context.Context, am AuthMethod) error {
	if am == nil {
		return errors.New("auth handler: nil auth method")
	}

	ah.logger.Info("starting auth handler")
	defer func() {
		am.Shutdown()
		/*close(ah.OutputCh)
		close(ah.TemplateTokenCh)*/
		ah.logger.Info("auth handler stopped")
	}()

	credCh := am.NewCreds()
	if !ah.enableReauthOnNewCredentials {
		realCredCh := credCh
		credCh = nil
		if realCredCh != nil {
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					case <-realCredCh:
					}
				}
			}()
		}
	}
	if credCh == nil {
		credCh = make(chan struct{})
	}

	var watcher *vaultapi.LifetimeWatcher

	for {
		select {
		case <-ctx.Done():
			return nil

		default:
		}

		// Create a fresh backoff value
		backoff := 2*time.Second + time.Duration(ah.random.Int63()%int64(time.Second*2)-int64(time.Second))

		path, header, data, err := am.Authenticate(ctx, ah.client)

		if err != nil {
			ah.logger.Error(err, "Error getting path or data from method", "backoff", backoff.Seconds())
			return err
		}

		var clientToUse *vaultapi.Client

		switch am.(type) {
		case AuthMethodWithClient:
			clientToUse, err = am.(AuthMethodWithClient).AuthClient(ah.client)
			if err != nil {
				ah.logger.Error(err, "Error creating client for authentication call", "backoff", backoff.Seconds())
				return err
			}
		default:
			clientToUse = ah.client
		}

		for key, values := range header {
			for _, value := range values {
				clientToUse.AddHeader(key, value)
			}
		}

		secret, err := clientToUse.Logical().Write(path, data)

		// Check errors/sanity
		if err != nil {
			ah.logger.Error(err, "Error authenticating", "backoff", backoff.Seconds())
			return err
		}

		if secret == nil || secret.Auth == nil {
			ah.logger.Error(err, "Authentication returned nil auth info", "backoff", backoff.Seconds())
			return err
		}

		if secret.Auth.ClientToken == "" {
			ah.logger.Error(err, "Authentication returned empty client token", "backoff", backoff.Seconds())
			return err
		}

		ah.logger.Info("Authentication successful")
		ah.client.SetToken(secret.Auth.ClientToken)
		am.CredSuccess()

		if watcher != nil {
			watcher.Stop()
		}

		watcher, err = clientToUse.NewLifetimeWatcher(&vaultapi.LifetimeWatcherInput{
			Secret: secret,
		})

		if err != nil {
			ah.logger.Error(err, "Error creating lifetime watcher, backing off and retrying", "backoff", backoff.Seconds())
			return err
		}

		// Start the renewal process
		ah.logger.Info("Starting renewal process")
		go watcher.Renew()

	LifetimeWatcherLoop:
		for {
			select {
			case <-ctx.Done():
				ah.logger.Info("Shutdown triggered, stopping lifetime watcher")
				watcher.Stop()
				break LifetimeWatcherLoop

			case err := <-watcher.DoneCh():
				ah.logger.Info("Lifetime watcher done channel triggered")
				if err != nil {
					ah.logger.Error(err, "Error renewing token")
				}
				break LifetimeWatcherLoop

			case <-watcher.RenewCh():
				ah.logger.Info("Renewed auth token")

			case <-credCh:
				ah.logger.Info("Auth method found new credentials, re-authenticating")
				break LifetimeWatcherLoop
			}
		}
	}
}
