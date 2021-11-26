package vault

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/hashicorp/vault/api"
	. "github.com/onsi/gomega"
)

type testAuthHandler struct {
	path   string
	header http.Header
	data   map[string]interface{}
	err    error
}

func (h *testAuthHandler) Authenticate(ctx context.Context) (string, http.Header, map[string]interface{}, error) {
	return h.path, h.header, h.data, h.err
}

type testTokenWriter struct {
	token string
}

func (w *testTokenWriter) SetToken(token string) {
	w.token = token
}

func TestAuthenticate(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name            string
		readWriter      *mockReadWriter
		tokenWriter     *testTokenWriter
		authHandler     *testAuthHandler
		expectToken     string
		expectLoginPath string
		expectError     error
	}{
		/*{
			name:        "errors if auth handler is nil",
			authHandler: nil,
			tokenWriter: &testTokenWriter{},
			readWriter: &mockReadWriter{
				writeResult: testResult{
					err: nil,
				},
			},
			expectError: errors.New("auth handler: nil auth method"),
		},*/
		{
			name: "fails if auth handler fails",
			authHandler: &testAuthHandler{
				err: errors.New("auth handler failed"),
			},
			tokenWriter: &testTokenWriter{},
			readWriter: &mockReadWriter{
				writeResult: testResult{
					err: nil,
				},
			},
			expectError: errors.New("error getting path or data from method: auth handler failed"),
		},
		{
			name: "fails if auth request fails",
			authHandler: &testAuthHandler{
				path: "/auth/dummy",
			},
			tokenWriter: &testTokenWriter{},
			readWriter: &mockReadWriter{
				writeResult: testResult{
					err: errors.New("auth request failed"),
				},
			},
			expectError: errors.New("login request failed: auth request failed"),
		},
		{
			name: "fails if auth request response is nil",
			authHandler: &testAuthHandler{
				path: "/auth/dummy",
			},
			tokenWriter: &testTokenWriter{},
			readWriter: &mockReadWriter{
				writeResult: testResult{},
			},
			expectError: errors.New("authentication returned nil auth info"),
		},
		{
			name: "fails if auth response contains no Auth information",
			authHandler: &testAuthHandler{
				path: "/auth/dummy",
			},
			tokenWriter: &testTokenWriter{},
			readWriter: &mockReadWriter{
				writeResult: testResult{
					secret: &api.Secret{},
				},
			},
			expectError: errors.New("authentication returned nil auth info"),
		},
		{
			name: "fails if auth response contains no auth token",
			authHandler: &testAuthHandler{
				path: "/auth/dummy",
			},
			tokenWriter: &testTokenWriter{},
			readWriter: &mockReadWriter{
				writeResult: testResult{
					secret: &api.Secret{
						Auth: &api.SecretAuth{},
					},
				},
			},
			expectError: errors.New("authentication returned empty client token"),
		},
		{
			name: "Set token if auth was successful",
			authHandler: &testAuthHandler{
				path: "/auth/dummy",
			},
			tokenWriter: &testTokenWriter{},
			readWriter: &mockReadWriter{
				writeResult: testResult{
					secret: &api.Secret{
						Auth: &api.SecretAuth{
							ClientToken: "banana",
						},
					},
				},
			},
			expectError: nil,
			expectToken: "banana",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewAuthHandler(AuthHandlerConfig{
				Writer:      test.readWriter,
				TokenWriter: test.tokenWriter,
			})

			err := handler.Authenticate(context.TODO(), test.authHandler)
			if test.expectError == nil {
				g.Expect(err).NotTo(HaveOccurred(), "write error occurd but should not")
			} else {
				g.Expect(err.Error()).To(Equal(test.expectError.Error()))
			}

			g.Expect(test.tokenWriter.token).To(Equal(test.expectToken))
			g.Expect(test.readWriter.writtenPath).To(Equal(test.authHandler.path))
		})
	}
}
