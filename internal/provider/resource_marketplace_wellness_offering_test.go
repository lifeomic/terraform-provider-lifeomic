package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/lifeomic/terraform-provider-lifeomic/internal/common"
)

func TestAccMarketplaceWellnessOffering_basic(t *testing.T) {
	t.Parallel()

	id, _ := uuid.GenerateUUID()
	err := os.Setenv(common.HeadersEnvVar, "{\"LifeOmic-Policy\":\"{\\\"rules\\\":{\\\"publishContent\\\":true}}\",\"LifeOmic-User\":\"wellness-service\",\"LifeOmic-Account\":\"lifeomic\"}")
	if err != nil {
		t.Fatalf("error setting required headers %v", err)
	}
	header, err := common.HeaderFromEnv()
	if err != nil {
		t.Fatalf("error getting required headers %v", err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories,

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				  resource "lifeomic_marketplace_wellness_offering" "test" {
						id = "%s"
						title = "Fake Module"
						description = "A fake marketplace module"
						marketplace_provider = "LifeOmic"
						version = "1.0.0"
						image_url = "https://placekitten.com/1800/1600"
						info_url = "https://example.com"
						approximate_unit_cost = 10000
						configuration_schema = jsonencode({
							"version": "06-28-2021",
							"fields": []
							})
						is_enabled = true
						install_url = "lambda://wellness-service:deployed/v1/private/life-league"
						is_test_module = true
					}
				`, id),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						client := newClientSet("", "", header).Marketplace

						_, err := client.GetPublishedModule(context.Background(), id, "")
						if err != nil {
							return err
						}

						return nil
					},
				),
			},
		},
	})
}
