package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/lifeomic/terraform-provider-lifeomic/internal/common"
)

const (
	defaultHeadersf = "{\"LifeOmic-Policy\":\"{\\\"rules\\\":{\\\"publishContent\\\":true, \\\"lifeomicMarketplaceAdmin\\\": true}}\",\"LifeOmic-User\":\"%s\",\"LifeOmic-Account\":\"%s\"}"
)

func getHeaders(t *testing.T) string {
	t.Helper()

	username := "tf-provider"
	account := "lifeomic"

	return fmt.Sprintf(defaultHeadersf, username, account)
}

var testWellnessOfferingResName = "lifeomic_marketplace_wellness_offering.test"

func TestAccMarketplaceWellnessOffering_basic(t *testing.T) {
	id, _ := uuid.GenerateUUID()
	t.Setenv(common.HeadersEnvVar, getHeaders(t))
	t.Setenv(useLambdaEnvVar, "1")
	header, err := common.HeaderFromEnv()
	if err != nil {
		t.Fatalf("error getting required headers %v", err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories,

		Steps: []resource.TestStep{
			{
				Config: testAccOffering_basic(id, true, "1.0.0"),
				Check: resource.ComposeAggregateTestCheckFunc(testCheckPublishedModule(t, id, header),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "is_test_module", "true"),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "is_approved", "true"),
				),
			},
		},
	})
}

func TestAccMarketplaceWellnessOffering_basicUpdate(t *testing.T) {
	id, _ := uuid.GenerateUUID()

	t.Setenv(common.HeadersEnvVar, getHeaders(t))
	t.Setenv(useLambdaEnvVar, "1")
	header, err := common.HeaderFromEnv()
	if err != nil {
		t.Fatalf("error getting required headers %v", err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories,

		Steps: []resource.TestStep{
			{
				Config: testAccOffering_basic(id, true, "1.0.0"),
				Check: resource.ComposeAggregateTestCheckFunc(testCheckPublishedModule(t, id, header),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "is_approved", "true"),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "is_test_module", "true")),
			},
			{
				Config: testAccOffering_basic(id, true, "1.0.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						client := newClientSet("", "", header).Marketplace

						module, err := client.GetPublishedModule(context.Background(), id, "")
						if err != nil {
							return err
						}

						if module.MyModule.Version != "1.0.1" {
							t.Fatalf("expected module version to be 1.0.1, instead got %s", module.MyModule.Version)
						}

						return nil
					},
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "is_approved", "true"),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "is_test_module", "true"),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "version", "1.0.1"),
				),
			},
		},
	})
}

func TestAccMarketplaceWellnessOffering_automaticApproval(t *testing.T) {
	id, _ := uuid.GenerateUUID()

	t.Setenv(common.HeadersEnvVar, getHeaders(t))
	t.Setenv(useLambdaEnvVar, "1")
	header, err := common.HeaderFromEnv()
	if err != nil {
		t.Fatalf("error getting required headers %v", err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories,

		Steps: []resource.TestStep{
			{
				Config: testAccOffering_basic(id, false, "1.0.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						client := newClientSet("", "", header).Marketplace

						_, err := client.GetPublishedModule(context.Background(), id, "")
						if err != nil {
							return err
						}

						return nil
					},
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "is_approved", "true"),
				),
			},
		},
	})
}

func TestAccMarketplaceWellnessOffering_automaticApprovalWithUpdates(t *testing.T) {
	id, _ := uuid.GenerateUUID()

	t.Setenv(common.HeadersEnvVar, getHeaders(t))
	t.Setenv(useLambdaEnvVar, "1")
	header, err := common.HeaderFromEnv()
	if err != nil {
		t.Fatalf("error getting required headers %v", err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories,

		Steps: []resource.TestStep{
			{
				Config: testAccOffering_basic(id, false, "1.0.0"),
				Check: resource.ComposeAggregateTestCheckFunc(testCheckPublishedModule(t, id, header),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "is_approved", "true")),
			},
			{
				Config: testAccOffering_basic(id, false, "1.0.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						client := newClientSet("", "", header).Marketplace

						module, err := client.GetPublishedModule(context.Background(), id, "")
						if err != nil {
							return err
						}

						if module.MyModule.Version != "1.0.1" {
							t.Fatalf("expected module version to be 1.0.1, instead got %s", module.MyModule.Version)
						}

						return nil
					},
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "is_approved", "true"),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "version", "1.0.1"),
				),
			},
		},
	})
}

// TODO: finalize this test.
// - require valid PHC account credentials

// func TestAccMarketplaceWellnessOffering_noAutoApproval(t *testing.T) {
// 	id, _ := uuid.GenerateUUID()

// 	t.Setenv(common.HeadersEnvVar, getHeaders(t))
// 	t.Setenv(useLambdaEnvVar, "")
// 	header, err := common.HeaderFromEnv()
// 	if err != nil {
// 		t.Fatalf("error getting required headers %v", err)
// 	}

// 	resource.Test(t, resource.TestCase{
// 		ProtoV6ProviderFactories: testAccProviderFactories,

// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccOffering_basic(id, false),
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					func(s *terraform.State) error {
// 						client := newClientSet("", "", header).Marketplace

// 						_, err := client.GetDraftWellnessOfferingModule(context.Background(), id)
// 						if err != nil {
// 							return err
// 						}

// 						return nil
// 					},
// 					resource.TestCheckResourceAttr(testWellnessOfferingResName, "is_approved", "false"),
// 				),
// 			},
// 		},
// 	})
// }

func testAccOffering_basic(id string, isTest bool, version string) string {
	return fmt.Sprintf(`resource "lifeomic_marketplace_wellness_offering" "test" {
	id = "%s"
	title = "Fake Module"
	description = "A fake marketplace module"
	marketplace_provider = "LifeOmic"
	version = "%s"
	image_url = "https://placekitten.com/1800/1600"
	info_url = "https://example.com"
	approximate_unit_cost = 10000
	configuration_schema = jsonencode({
		"version": "06-28-2021",
		"fields": []
		})
	is_enabled = true
	install_url = "lambda://wellness-service:deployed/v1/private/life-league"
	is_test_module = %t
	}`, id, version, isTest)
}

func testCheckPublishedModule(t *testing.T, id string, header map[string]string) func(s *terraform.State) error {
	t.Helper()
	return func(s *terraform.State) error {

		client := newClientSet("", "", header).Marketplace

		_, err := client.GetPublishedModule(context.Background(), id, "")
		if err != nil {
			return err
		}

		return nil
	}
}
