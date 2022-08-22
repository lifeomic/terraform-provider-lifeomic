package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/lifeomic/terraform-provider-lifeomic/internal/client"
)

const useLambdaEnvVar = "LIFEOMIC_USE_LAMBDA"

type provider struct {
	clientSet  *clientSet
	configured bool
}

type providerData struct {
	AccountID types.String `tfsdk:"account_id"`
	Token     types.String `tfsdk:"token"`
	Host      types.String `tfsdk:"host"`
}

func New() tfsdk.Provider {
	return &provider{}
}

func providerAttributeDescription(description, envVar string) string {
	return fmt.Sprintf("%s. If not set explicitly in the provider block, `$%s` will be used.", description, envVar)
}

func (p *provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"account_id": {
				Type:        types.StringType,
				Optional:    true,
				Description: providerAttributeDescription("The unique ID of the PHC Account to use this provider with", client.AccountIDEnvVar),
			},
			"token": {
				Type:        types.StringType,
				Sensitive:   true,
				Optional:    true,
				Description: providerAttributeDescription("The token to use for authenticating with the PHC API", client.AuthTokenEnvVar),
			},
			"host": {
				Type:        types.StringType,
				Optional:    true,
				Description: providerAttributeDescription("The PHC API host to communicate with.", client.HostEnvVar),
			},
		},
	}, nil
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	tflog.Trace(ctx, "Configuring provider")
	config := new(providerData)

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	requireProviderValue(resp, "token", client.AuthTokenEnvVar, &config.Token)
	requireProviderValue(resp, "account_id", client.AccountIDEnvVar, &config.AccountID)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Provider configuration", map[string]any{
		"provider": p,
	})

	p.clientSet = newClientSet(config.Token.Value, config.AccountID.Value)
	p.configured = true
}

func requireProviderValue(resp *tfsdk.ConfigureProviderResponse, attribute, envVar string, value *types.String) {
	if value.Value != "" {
		return
	}

	val, ok := os.LookupEnv(envVar)
	if !ok {
		resp.Diagnostics.AddAttributeError(path.Root(attribute),
			fmt.Sprintf("Missing required provider value %q", attribute),
			fmt.Sprintf("Either set %q in the provider block or via the %s environment variable", attribute, envVar))
		return
	}

	value.Value = val
}

func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"lifeomic_policy":                        policyResourceType{},
		"lifeomic_marketplace_wellness_offering": wellnessOfferingResourceType{},
	}, nil
}

func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{}, nil
}

func errorConvertingProvider(v any) diag.Diagnostics {
	return diag.Diagnostics{diag.NewErrorDiagnostic("Error converting provider",
		fmt.Sprintf("An unexpected error was encountered converting the provider."+
			"This is always a bug in the provider.\n\nType: %T", v))}
}
