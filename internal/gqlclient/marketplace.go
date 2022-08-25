package gqlclient

// Generated by ./cmd/service-client-gen

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/lifeomic/terraform-provider-lifeomic/internal/client"
)

const (
	marketplaceServiceName     = "marketplace-service"
	marketplaceDefaultEndpoint = "https://marketplace.us.lifeomic.com/v1/marketplace/authenticated/graphql"
)

type MarketplaceService interface {
	GetPublishedModule(ctx context.Context, id string, version string) (*GetPublishedModuleResponse, error)
	CreateDraftModule(ctx context.Context, input CreateDraftModuleInput) (*CreateDraftModuleResponse, error)
	DeleteModule(ctx context.Context, input DeleteModuleInput) (*DeleteModuleResponse, error)
	SetAppTile(ctx context.Context, input SetPublicAppTileDraftModuleSourceInput) (*SetAppTileResponse, error)
	PublishModule(ctx context.Context, input PublishDraftModuleInputV2) (*PublishModuleResponse, error)
	PublishModuleV3(ctx context.Context, input PublishDraftModuleInputV3) (*PublishModuleV3Response, error)
	StartImageUpload(ctx context.Context, input StartUploadInput) (*StartImageUploadResponse, error)
	FinalizeImageUpload(ctx context.Context, input FinalizeUploadInput) (*FinalizeImageUploadResponse, error)
	SetWellnessOfferingDraftModuleSource(ctx context.Context, input SetDraftModuleWellnessOfferingSourceInput) (*SetWellnessOfferingDraftModuleSourceResponse, error)
	GetWellnessOfferingModule(ctx context.Context, moduleId string) (*GetWellnessOfferingModuleResponse, error)
	UpdateDraftModule(ctx context.Context, input UpdateDraftModuleInput) (*UpdateDraftModuleResponse, error)
}

type marketplaceClient struct {
	client graphql.Client
}

func (m *marketplaceClient) GetPublishedModule(ctx context.Context, id string, version string) (*GetPublishedModuleResponse, error) {
	return GetPublishedModule(ctx, m.client, id, version)
}

func (m *marketplaceClient) CreateDraftModule(ctx context.Context, input CreateDraftModuleInput) (*CreateDraftModuleResponse, error) {
	return CreateDraftModule(ctx, m.client, input)
}

func (m *marketplaceClient) DeleteModule(ctx context.Context, input DeleteModuleInput) (*DeleteModuleResponse, error) {
	return DeleteModule(ctx, m.client, input)
}

func (m *marketplaceClient) SetAppTile(ctx context.Context, input SetPublicAppTileDraftModuleSourceInput) (*SetAppTileResponse, error) {
	return SetAppTile(ctx, m.client, input)
}

func (m *marketplaceClient) PublishModule(ctx context.Context, input PublishDraftModuleInputV2) (*PublishModuleResponse, error) {
	return PublishModule(ctx, m.client, input)
}

func (m *marketplaceClient) PublishModuleV3(ctx context.Context, input PublishDraftModuleInputV3) (*PublishModuleV3Response, error) {
	return PublishModuleV3(ctx, m.client, input)
}

func (m *marketplaceClient) StartImageUpload(ctx context.Context, input StartUploadInput) (*StartImageUploadResponse, error) {
	return StartImageUpload(ctx, m.client, input)
}

func (m *marketplaceClient) FinalizeImageUpload(ctx context.Context, input FinalizeUploadInput) (*FinalizeImageUploadResponse, error) {
	return FinalizeImageUpload(ctx, m.client, input)
}

func (m *marketplaceClient) SetWellnessOfferingDraftModuleSource(ctx context.Context, input SetDraftModuleWellnessOfferingSourceInput) (*SetWellnessOfferingDraftModuleSourceResponse, error) {
	return SetWellnessOfferingDraftModuleSource(ctx, m.client, input)
}

func (m *marketplaceClient) GetWellnessOfferingModule(ctx context.Context, moduleId string) (*GetWellnessOfferingModuleResponse, error) {
	return GetWellnessOfferingModule(ctx, m.client, moduleId)
}

func (m *marketplaceClient) UpdateDraftModule(ctx context.Context, input UpdateDraftModuleInput) (*UpdateDraftModuleResponse, error) {
	return UpdateDraftModule(ctx, m.client, input)
}

func NewMarketplaceClient(authToken string, accountID string, header map[string]string) MarketplaceService {
	transport := client.NewAuthedTransport(authToken, accountID, marketplaceServiceName, header)
	return &marketplaceClient{client: graphql.NewClient(marketplaceDefaultEndpoint, transport)}
}
