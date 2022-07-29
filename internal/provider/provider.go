package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/lifeomic/terraform-provider-phc/internal/client"
)

func New() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The PHC API host to configure the client to use.",
			},
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("PHC_TOKEN", ""),
				Description: "The token to use for authenticating with the PHC API. If not explicitly set, it will be sourced from the PHC_TOKEN environment variable.",
			},
			"account_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique identifier of the LifeOmic account to use when communicating with the API.",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{},
		ResourcesMap: map[string]*schema.Resource{
			"phc_policy": resourcePolicy(),
		},

		ConfigureContextFunc: configureProvider,
	}
}

type providerMeta struct {
	Client client.Interface
}

func configureProvider(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	config := client.Config{
		Host:      d.Get("host").(string),
		AuthToken: d.Get("token").(string),
		Account:   d.Get("account_id").(string),
	}
	return &providerMeta{Client: client.New(config)}, nil
}
