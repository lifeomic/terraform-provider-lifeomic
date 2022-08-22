package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/lifeomic/terraform-provider-lifeomic/internal/gqlclient"
)

// wellnessOffering represents the state of marketplace_wellness_offering resource
type wellnessOffering struct {
	ID                  types.String `tfsdk:"id"`
	ParentModuleId      types.String `tfsdk:"parent_module_id"`
	Title               types.String `tfsdk:"title"`
	Description         types.String `tfsdk:"description"`
	MarketplaceProvider types.String `tfsdk:"marketplace_provider"`
	Version             types.String `tfsdk:"version"`
	ImageURL            types.String `tfsdk:"image_url"`
	InfoURL             types.String `tfsdk:"info_url"`
	ApproximateUnitCost types.Int64  `tfsdk:"approximate_unit_cost"`
	InstallURL          types.String `tfsdk:"install_url"`
	ConfigurationSchema types.String `tfsdk:"configuration_schema"`
	IsEnabled           types.Bool   `tfsdk:"is_enabled"`
	IsTestModule        types.Bool   `tfsdk:"is_test_module"`
}

// wellnessOfferingResource implements tfsdk
type wellnessOfferingResource struct {
	clientSet *clientSet
}

// wellnessOfferingResourceType implements tfsdk.ResourceType
type wellnessOfferingResourceType struct{}

func (wellnessOfferingResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "marketplace_wellness_offering manages Wellness Offering subsidies",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Optional:    true,
				Type:        types.StringType,
				Description: "An optional id for the Wellness Offering",
			},
			"parent_module_id": {
				Optional: true,
				Type:     types.StringType,
			},
			"title": {
				Required:    true,
				Type:        types.StringType,
				Description: "The title of the Wellness Offering",
			},
			"description": {
				Required:    true,
				Type:        types.StringType,
				Description: "The description of the Wellness Offering",
			},
			"marketplace_provider": {
				Required: true,
				Type:     types.StringType,
			},
			"version": {
				Required: true,
				Type:     types.StringType,
			},
			"image_url": {
				Required: true,
				Type:     types.StringType,
			},
			"info_url": {
				Required: true,
				Type:     types.StringType,
			},
			"approximate_unit_cost": {
				Required:    true,
				Type:        types.Int64Type,
				Description: "The approximate per unit cost of the module represented in USD Pennies",
			},
			"install_url": {
				Required: true,
				Type:     types.StringType,
			},
			"configuration_schema": {
				Required: true,
				Type:     types.StringType,
			},
			"is_enabled": {
				Required: true,
				Type:     types.BoolType,
			},
			"is_test_module": {
				Optional: true,
				Type:     types.BoolType,
			},
		},
	}, nil
}

func (w wellnessOffering) ToMarketplaceInputObject(ctx context.Context) (gqlclient.CreateDraftModuleInput, diag.Diagnostics) {
	return gqlclient.CreateDraftModuleInput{
		Category:       "WELLNESS_OFFERING",
		Description:    w.Description.Value,
		Id:             w.ID.Value,
		Title:          w.Title.Value,
		ParentModuleId: w.ParentModuleId.Value,
		Icon:           nil,
	}, nil
}

func (w wellnessOffering) ToMarketplaceUpdateInput(ctx context.Context) (gqlclient.UpdateDraftModuleInput, diag.Diagnostics) {
	return gqlclient.UpdateDraftModuleInput{
		Description:    w.Description.Value,
		Title:          w.Title.Value,
		ModuleId:       w.ID.Value,
		ParentModuleId: w.ParentModuleId.Value,
	}, nil
}

func (wellnessOfferingResourceType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	pr, ok := p.(*provider)
	if !ok {
		return nil, errorConvertingProvider(p)
	}

	return &wellnessOfferingResource{
		clientSet: pr.clientSet,
	}, nil
}

func (w wellnessOfferingResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	tflog.Info(ctx, "Creating Wellness Offering Module")

	// Get plan values.
	var plan wellnessOffering
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map Terraform plan to *client.WellnessOffering object.
	draftModuleInput, diags := plan.ToMarketplaceInputObject(ctx)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Create the draft module.
	draftModuleResp, err := w.clientSet.Marketplace.CreateDraftModule(ctx, draftModuleInput)
	if err != nil {
		resp.Diagnostics.AddError("failed to create Wellness Draft Module", err.Error())
		return
	}

	tflog.Info(ctx, "Created new DraftModule", map[string]any{"draftModule": draftModuleResp.CreateDraftModule})

	// Set module source
	setSourceResp, err := w.clientSet.Marketplace.SetWellnessOfferingDraftModuleSource(ctx, gqlclient.SetDraftModuleWellnessOfferingSourceInput{
		ModuleId: draftModuleResp.CreateDraftModule.Id,
		SourceInfo: gqlclient.WellnessOfferingModuleSourceInfo{
			ApproximateUnitCost: int(plan.ApproximateUnitCost.Value),
			ConfigurationSchema: plan.ConfigurationSchema.Value,
			ImageUrl:            plan.ImageURL.Value,
			InfoUrl:             plan.InfoURL.Value,
			InstallUrl:          plan.InstallURL.Value,
			Provider:            plan.MarketplaceProvider.Value,
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to set source of Wellness Draft Module", err.Error())
		return
	}

	tflog.Info(ctx, "Set module source", map[string]any{"moduleSource": setSourceResp.SetWellnessOfferingDraftModuleSource})

	// Publish module
	publishResp, err := w.clientSet.Marketplace.PublishModuleV3(ctx, gqlclient.PublishDraftModuleInputV3{
		ModuleId: draftModuleResp.CreateDraftModule.Id,
		Version: gqlclient.ModuleVersionInput{
			Version: plan.Version.Value,
		},
		IsTestModule: plan.IsTestModule.Value,
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to publish Wellness Offering Module", err.Error())
		return
	}

	tflog.Info(ctx, "Published module", map[string]any{"module": publishResp.PublishDraftModuleV3})

	offering, err := w.clientSet.Marketplace.GetWellnessOfferingModule(ctx, publishResp.PublishDraftModuleV3.Id)
	if err != nil {
		resp.Diagnostics.AddError("failed to get published Wellness Offering Module", err.Error())
		return
	}

	tflog.Info(ctx, "Got Wellness Offering Module", map[string]any{"module": offering.MyModule.WellnessOfferingModule})
	resp.Diagnostics.Append(setWellnessOfferingState(ctx, &plan, &resp.State, offering.MyModule.WellnessOfferingModule)...)
}

func (w wellnessOfferingResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	tflog.Info(ctx, "Reading Wellness Offering resource")

	// Get current state.
	var state wellnessOffering
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	offering, err := w.clientSet.Marketplace.GetWellnessOfferingModule(ctx, state.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("failed to get wellness offering module", err.Error())
		return
	}

	tflog.Info(ctx, "Got Wellness Offering Module", map[string]any{"module": offering.MyModule})
	resp.Diagnostics.Append(setWellnessOfferingState(ctx, &state, &resp.State, offering.MyModule.WellnessOfferingModule)...)
}

func (w wellnessOfferingResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	tflog.Info(ctx, "Updating Wellness Offering Module")

	// Get plan values.
	var plan wellnessOffering
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state.
	var state wellnessOffering
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	draftModuleInput, diags := plan.ToMarketplaceUpdateInput(ctx)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	updateResp, err := w.clientSet.Marketplace.UpdateDraftModule(ctx, draftModuleInput)
	if err != nil {
		resp.Diagnostics.AddError("failed to update Wellness Offering draft module", err.Error())
		return
	}
	tflog.Info(ctx, "Updated existing Wellness Offering Module", map[string]any{"moduleUpdate": updateResp.UpdateDraftModule})

	setSourceResp, err := w.clientSet.Marketplace.SetWellnessOfferingDraftModuleSource(ctx, gqlclient.SetDraftModuleWellnessOfferingSourceInput{
		ModuleId: updateResp.UpdateDraftModule.Id,
		SourceInfo: gqlclient.WellnessOfferingModuleSourceInfo{
			ApproximateUnitCost: int(plan.ApproximateUnitCost.Value),
			ConfigurationSchema: plan.ConfigurationSchema.Value,
			ImageUrl:            plan.ImageURL.Value,
			InfoUrl:             plan.InfoURL.Value,
			InstallUrl:          plan.InstallURL.Value,
			Provider:            plan.MarketplaceProvider.Value,
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to set source of Wellness Offering draft module", err.Error())
		return
	}

	tflog.Info(ctx, "Updated existing Wellness Offering Module source", map[string]any{"moduleSource": setSourceResp.SetWellnessOfferingDraftModuleSource})

	offering, err := w.clientSet.Marketplace.GetWellnessOfferingModule(ctx, setSourceResp.SetWellnessOfferingDraftModuleSource.Id)
	if err != nil {
		resp.Diagnostics.AddError("failed to get wellness offering module", err.Error())
		return
	}

	resp.Diagnostics.Append(setWellnessOfferingState(ctx, &plan, &resp.State, offering.MyModule.WellnessOfferingModule)...)
}

func (w wellnessOfferingResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	tflog.Info(ctx, "Deleting Wellness Offering Module")

	// Get current state.
	var state wellnessOffering
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteModuleResp, err := w.clientSet.Marketplace.DeleteModule(ctx, gqlclient.DeleteModuleInput{
		ModuleId: state.ID.Value,
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to delete Wellness Offering Module", err.Error())
		return
	}

	tflog.Info(ctx, "deleted Wellness Offering", map[string]any{"deleteResp": deleteModuleResp.DeleteModule})

	resp.State.RemoveResource(ctx)
	tflog.Info(ctx, "Deleted Wellness Offering", map[string]any{"Name": state.Title})
}

func setWellnessOfferingState(ctx context.Context, config *wellnessOffering, state *tfsdk.State, w gqlclient.WellnessOfferingModule) (diags diag.Diagnostics) {

	source, ok := w.Source.(*gqlclient.WellnessOfferingModuleSourceWellnessOffering)
	if !ok {
		diags.AddError("expected module source to be a wellness module", w.Source.GetTypename())
		return
	}

	diags.Append(state.Set(ctx, wellnessOffering{
		ParentModuleId: config.ParentModuleId,
		IsEnabled:      config.IsEnabled,
		IsTestModule:   config.IsTestModule,
		InstallURL:     config.InstallURL,

		ID:                  types.String{Value: w.Id},
		Title:               types.String{Value: w.Title},
		Description:         types.String{Value: w.Description},
		MarketplaceProvider: types.String{Value: source.Provider},
		Version:             types.String{Value: w.Version},
		ImageURL:            types.String{Value: source.ImageUrl},
		InfoURL:             types.String{Value: source.InfoUrl},
		ApproximateUnitCost: types.Int64{Value: int64(source.ApproximateUnitCost)},
		ConfigurationSchema: types.String{Value: source.ConfigurationSchema},
	})...)
	return
}
