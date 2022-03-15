package main

import (
	"github.com/lifeomic/terraform-provider-phc/marketplace"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/lifeomic/phc-sdk-go/client"
)

func buildRules(rawRules map[string]interface{}) map[string]bool {
	rules := map[string]bool{}
	for key, val := range rawRules {
		if val == "true" {
			rules[key] = true
		} else {
			rules[key] = false
		}
	}
	return rules
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	rules := buildRules(d.Get("rules").(map[string]interface{}))
	return client.BuildClient("lifeomic", "phc-tf", rules)
}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		ConfigureFunc: providerConfigure,
		Schema: map[string]*schema.Schema{
			"rules": {
				Type:     schema.TypeMap,
				Optional: true,
				Default:  map[string]interface{}{},
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"app_tile": marketplace.AppTileResource(),
		},
	}
}
