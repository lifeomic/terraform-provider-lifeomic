package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testPolicyResName = "lifeomic_policy.test"

func TestAccPHCPolicy_basic(t *testing.T) {
	t.Parallel()
	name := randomResourceName(t, 8)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV6ProviderFactories: testAccProviderFactories,

		Steps: []resource.TestStep{
			{
				Config: testAccPHCPolicy_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkPolicyExists,
					resource.TestCheckResourceAttr(testPolicyResName, "name", name),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.0.operation", "readData"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.0.comparison.#", "1"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.0.comparison.0.type", "includes"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.0.comparison.0.value", "admin"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.0.comparison.0.subject", "user.groups"),
				),
			},
			{
				Config:             testAccPHCPolicy_basic(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccPHCPolicy_staticRule(t *testing.T) {
	t.Parallel()
	name := randomResourceName(t, 8)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV6ProviderFactories: testAccProviderFactories,

		Steps: []resource.TestStep{
			{
				Config: testAccPHCPolicy_staticRule(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkPolicyExists,
					resource.TestCheckResourceAttr(testPolicyResName, "name", name),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.0.operation", "readData"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.0.allowed", "true"),
				),

				// Unfortunately something about the schema is causing Terraform to produce
				// a *empty* plan to update-in-place after every apply where `allowed` is not
				// null. This is probably an upstream bug with the new framework-sdk that
				// needs more investigation and an issue to be created (probably has something
				// to do with `allowed` being a pointer).
				//
				// TODO: remove this flag when fixed
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccPHCPolicy_update(t *testing.T) {
	t.Parallel()
	name := randomResourceName(t, 8)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV6ProviderFactories: testAccProviderFactories,

		Steps: []resource.TestStep{
			{
				Config: testAccPHCPolicy_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkPolicyExists,
					resource.TestCheckResourceAttr(testPolicyResName, "name", name),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.0.operation", "readData"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.0.comparison.#", "1"),
				),
			},

			{
				Config: testAccPHCPolicy_manyRules(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkPolicyExists,
					resource.TestCheckResourceAttr(testPolicyResName, "name", name),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.#", "3"),

					resource.TestCheckResourceAttr(testPolicyResName, "rule.0.operation", "readMaskedData"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.0.comparison.#", "2"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.0.comparison.0.type", "in"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.0.comparison.0.subject", "user.patients"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.0.comparison.0.target", "resource.subject"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.0.comparison.1.type", "equals"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.0.comparison.1.subject", "user.id"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.0.comparison.1.value", "bob"),

					resource.TestCheckResourceAttr(testPolicyResName, "rule.1.operation", "readData"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.1.comparison.#", "2"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.1.comparison.0.type", "includes"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.1.comparison.0.subject", "user.groups"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.1.comparison.0.value", "admin"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.1.comparison.1.type", "notEquals"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.1.comparison.1.subject", "resource.dataset"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.1.comparison.1.value", "eed10b7f-8b8d-4182-a10b-7bf541ed4e36"),

					resource.TestCheckResourceAttr(testPolicyResName, "rule.2.operation", "writeData"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.2.comparison.#", "1"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.2.comparison.0.type", "subset"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.2.comparison.0.values.#", "2"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.2.comparison.0.values.0", "admin"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.2.comparison.0.values.1", "doctor"),
					resource.TestCheckResourceAttr(testPolicyResName, "rule.2.comparison.0.subject", "user.groups"),
				),
			},
			{
				Config:             testAccPHCPolicy_manyRules(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccPHCPolicy_duplicateRuleBlock(t *testing.T) {
	t.Parallel()
	name := randomResourceName(t, 8)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV6ProviderFactories: testAccProviderFactories,

		Steps: []resource.TestStep{
			{
				Config:      testAccPHCPolicy_duplicateRuleBlock(name),
				ExpectError: regexp.MustCompile(`Duplicate occurance of rule for operation "readData"`),
			},
		},
	})
}

func TestAccPHCPolicy_conflictingRule(t *testing.T) {
	t.Parallel()
	name := randomResourceName(t, 8)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV6ProviderFactories: testAccProviderFactories,

		Steps: []resource.TestStep{
			{
				Config:      testAccPHCPolicy_conflictingRule(name),
				ExpectError: regexp.MustCompile("Exactly one of comparison and allowed should be set"),
			},
		},
	})
}

func TestAccPHCPolicy_invalidAllowed(t *testing.T) {
	t.Parallel()
	name := randomResourceName(t, 8)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV6ProviderFactories: testAccProviderFactories,

		Steps: []resource.TestStep{
			{
				Config:      testAccPHCPolicy_invalidAllowed(name),
				ExpectError: regexp.MustCompile("allowed must either be true or null"),
			},
		},
	})
}

func TestAccPHCPolicy_conflictingComparisonFields(t *testing.T) {
	t.Parallel()
	name := randomResourceName(t, 8)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV6ProviderFactories: testAccProviderFactories,

		Steps: []resource.TestStep{
			{
				Config:      testAccPHCPolicy_conflictingComparisonFields(name),
				ExpectError: regexp.MustCompile("Exactly one of value, values, or target must be set"),
			},
		},
	})
}

func checkPolicyExists(s *terraform.State) error {
	policyClient := newClientSet("", "", nil).Policies

	for _, res := range s.RootModule().Resources {
		if res.Type != "lifeomic_policy" {
			continue
		}

		if _, err := policyClient.Get(context.Background(), res.Primary.Attributes["name"]); err != nil {
			return err
		}
		break
	}
	return nil
}

func testAccPHCPolicy_basic(name string) string {
	return fmt.Sprintf(`resource "lifeomic_policy" "test" {
  name = "%s"
  
  rule {
    operation = "readData"

    comparison {
      subject = "user.groups"
      type    = "includes"
      value   = "admin"
    }
  }
}`, name)
}

func testAccPHCPolicy_staticRule(name string) string {
	return fmt.Sprintf(`resource "lifeomic_policy" "test" {
  name = "%s"
  
  rule {
    operation = "readData"
    allowed   = true
  }
}`, name)
}

func testAccPHCPolicy_conflictingRule(name string) string {
	return fmt.Sprintf(`resource "lifeomic_policy" "test" {
  name = "%s"
  
  rule {
    operation = "readData"

    allowed = true

    comparison {
      subject = "user.groups"
      type    = "includes"
      value   = "admin"
    }
  }
}`, name)
}

func testAccPHCPolicy_invalidAllowed(name string) string {
	return fmt.Sprintf(`resource "lifeomic_policy" "test" {
  name = "%s"
  
  rule {
    operation = "readData"
    allowed   = false
  }
}`, name)
}

func testAccPHCPolicy_conflictingComparisonFields(name string) string {
	return fmt.Sprintf(`resource "lifeomic_policy" "test" {
  name = "%s"
  
  rule {
    operation = "readData"

    comparison {
      subject = "user.groups"
      type    = "includes"
      value   = "admin"
      values  = ["admin"]
      target  = "resource.id"
    }
  }
}`, name)
}

func testAccPHCPolicy_duplicateRuleBlock(name string) string {
	return fmt.Sprintf(`resource "lifeomic_policy" "test" {
  name = "%s"
  
  rule {
    operation = "readData"

    comparison {
      subject = "user.groups"
      type    = "includes"
      value   = "admin"
    }
  }

  rule {
    operation = "readData"

    comparison {
      subject = "user.groups"
      type    = "includes"
      value   = "doctor"
    }
  }
}`, name)
}

func testAccPHCPolicy_manyRules(name string) string {
	return fmt.Sprintf(`resource "lifeomic_policy" "test" {
  name = "%s"
  rule {
    operation = "readMaskedData"

    comparison {
      subject = "user.patients"
      type    = "in"
      target  = "resource.subject"
    }

    comparison {
      subject = "user.id"
      type    = "equals"
      value   = "bob"
    }
  }
 
  rule {
    operation = "readData"

    comparison {
      subject = "user.groups"
      type    = "includes"
      value   = "admin"
    }

    comparison {
      subject = "resource.dataset"
      type    = "notEquals"
      value   = "eed10b7f-8b8d-4182-a10b-7bf541ed4e36"
    }
  }

  rule {
    operation = "writeData"

    comparison {
      subject = "user.groups"
      type    = "subset"
      values  = ["admin", "doctor"]
    }
  }
}`, name)
}
