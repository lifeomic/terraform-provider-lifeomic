package main

import (
	"github.com/lifeomic/terraform-provider-phc/appstore"
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
	return client.BuildClient(d.Get("account").(string), d.Get("user").(string), rules)
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
			"account": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "phc-tf",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"phc_app_tile": marketplace.AppTileResource(),
			"phc_applet":   appstore.AppletResource(),
		},
	}
}
