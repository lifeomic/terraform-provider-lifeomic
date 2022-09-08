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

const (
	defaultHeadersf = "{\"LifeOmic-Policy\":\"{\\\"rules\\\":{\\\"publishContent\\\":true, \\\"lifeomicMarketplaceAdmin\\\": true}}\",\"LifeOmic-User\":\"%s\",\"LifeOmic-Account\":\"%s\"}"
	defaultDesc     = "A fake marketplace wellness offering"
)

func skipNoLambda(t *testing.T) {
	if os.Getenv(useLambdaEnvVar) == "" {
		t.Skipf("skipping test. Set %s env var in order to run this test", useLambdaEnvVar)
	}
}

func getHeaders(t *testing.T) string {
	t.Helper()

	username := "tf-provider"
	account := "lifeomic"

	return fmt.Sprintf(defaultHeadersf, username, account)
}

var testWellnessOfferingResName = "lifeomic_marketplace_wellness_offering.test"

func TestAccMarketplaceWellnessOffering_basic(t *testing.T) {
	skipNoLambda(t)
	id, _ := uuid.GenerateUUID()
	t.Setenv(common.HeadersEnvVar, getHeaders(t))
	header, err := common.HeaderFromEnv()
	if err != nil {
		t.Fatalf("error getting required headers %v", err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories,

		Steps: []resource.TestStep{
			{
				Config: testAccOffering_basic(id, true, defaultDesc),
				Check: resource.ComposeAggregateTestCheckFunc(testCheckPublishedModule(t, id, header),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "is_test_module", "true"),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "is_approved", "true"),
				),
			},
		},
	})
}

func TestAccMarketplaceWellnessOffering_basicUpdate(t *testing.T) {
	skipNoLambda(t)
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
				Config: testAccOffering_basic(id, true, defaultDesc),
				Check: resource.ComposeAggregateTestCheckFunc(testCheckPublishedModule(t, id, header),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "is_approved", "true"),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "is_test_module", "true"),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "version", "1.0.0")),
			},
			{
				Config: testAccOffering_basic(id, true, "a new description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						client := newClientSet("", "", header).Marketplace

						module, err := client.GetPublishedModule(context.Background(), id, "")
						if err != nil {
							return err
						}

						if module.MyModule.Version != "1.1.0" {
							t.Fatalf("expected module version to be 1.1.0, instead got %s", module.MyModule.Version)
						}

						return nil
					},
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "is_approved", "true"),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "is_test_module", "true"),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "version", "1.1.0"),
				),
			},
		},
	})
}

func TestAccMarketplaceWellnessOffering_automaticApproval(t *testing.T) {
	skipNoLambda(t)
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
				Config: testAccOffering_basic(id, false, defaultDesc),
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
	skipNoLambda(t)
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
				Config: testAccOffering_basic(id, false, defaultDesc),
				Check: resource.ComposeAggregateTestCheckFunc(testCheckPublishedModule(t, id, header),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "is_approved", "true"),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "version", "1.0.0")),
			},
			{
				Config: testAccOffering_basic(id, false, "a really fake module"),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						client := newClientSet("", "", header).Marketplace

						module, err := client.GetPublishedModule(context.Background(), id, "")
						if err != nil {
							return err
						}

						if module.MyModule.Version != "1.1.0" {
							t.Fatalf("expected module version to be 1.1.0, instead got %s", module.MyModule.Version)
						}

						return nil
					},
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "is_approved", "true"),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "version", "1.1.0"),
				),
			},
		},
	})
}

func TestAccMarketplaceWellnessOffering_import(t *testing.T) {
	skipNoLambda(t)
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
				Config: testAccOffering_basic(id, false, defaultDesc),
				Check: resource.ComposeAggregateTestCheckFunc(testCheckPublishedModule(t, id, header),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "is_approved", "true"),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "version", "1.0.0")),
			},
			{
				ImportState:   true,
				ImportStateId: id,
				ResourceName:  testWellnessOfferingResName,
				Config:        testAccOffering_basic(id, false, defaultDesc),
				Check: resource.ComposeAggregateTestCheckFunc(testCheckPublishedModule(t, id, header),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "is_approved", "true"),
					resource.TestCheckResourceAttr(testWellnessOfferingResName, "version", "1.0.0")),
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

func testAccOffering_basic(id string, isTest bool, desc string) string {
	return fmt.Sprintf(`resource "lifeomic_marketplace_wellness_offering" "test" {
	id = "%s"
	title = "Fake Module"
	description = "%s"
	marketplace_provider = "LifeOmic"
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
	}`, id, desc, isTest)
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
