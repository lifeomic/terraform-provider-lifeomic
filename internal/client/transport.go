package client

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/lifeomic/terraform-provider-lifeomic/internal/lambda"
)

func NewAuthedTransport(authToken, accountID, serviceName string) *AuthedTransport {
	transport := &AuthedTransport{
		AuthToken: authToken,
		AccountID: accountID,
	}

	if user, ok := os.LookupEnv("LIFEOMIC_USER"); ok {
		transport.UserID = user
	}

	if GetUseLambda() {
		lambdaTransport, err := lambda.NewRoundTripper(context.Background(), lambda.URI{
			Function: serviceName,
		})
		if err != nil {
			log.Fatalf("failed to create lambda transport: %s", err)
		}

		transport.Base = lambdaTransport
	}
	return transport
}

type AuthedTransport struct {
	AuthToken string
	AccountID string
	UserID    string

	Base http.RoundTripper
}

func (t *AuthedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.AuthToken))
	}
	if t.AccountID != "" {
		req.Header.Set("LifeOmic-Account", t.AccountID)
	}
	if t.UserID != "" {
		req.Header.Set("LifeOmic-User", t.UserID)
	}

	baseTransport := t.Base
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}
	return baseTransport.RoundTrip(req)
}

func (t *AuthedTransport) Do(req *http.Request) (*http.Response, error) {
	return t.RoundTrip(req)
}

func GetUseLambda() bool {
	useLambda, _ := strconv.ParseBool(os.Getenv("LIFEOMIC_USE_LAMBDA"))
	return useLambda
}
