package provider

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/lifeomic/terraform-provider-lifeomic/internal/client"
)

var testAccProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"lifeomic": func() (tfprotov6.ProviderServer, error) {
		prv := New()
		server := providerserver.NewProtocol6(prv)
		return server(), nil
	},
}

// testAccPreCheck returns a function called before running acceptance tests.
// Ensure
func testAccPreCheck(t *testing.T) func() {
	t.Helper()

	return func() {
		requireEnvVar(t, client.AuthTokenEnvVar)
		requireEnvVar(t, client.AccountIDEnvVar)
	}
}

func requireEnvVar(t *testing.T, s string) {
	t.Helper()

	if _, ok := os.LookupEnv(s); !ok {
		t.Fatalf("%s must be set", s)
	}
}

func randomResourceName(t *testing.T, chars int) string {
	t.Helper()

	randBytes := make([]byte, chars)
	if _, err := rand.Read(randBytes); err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("tf-test-%X", randBytes)
}
