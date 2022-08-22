package provider

import (
	"github.com/lifeomic/terraform-provider-lifeomic/internal/client"
	"github.com/lifeomic/terraform-provider-lifeomic/internal/gqlclient"
)

type clientSet struct {
	AppStore    gqlclient.AppStoreService
	Policies    client.PolicyService
	Marketplace gqlclient.MarketplaceService
}

func newClientSet(token, accountID string) *clientSet {
	policiesClient := client.New(client.Config{
		AuthToken:   token,
		AccountID:   accountID,
		ServiceName: "account-service",
	}).Policies()

	return &clientSet{
		AppStore:    gqlclient.NewAppStoreClient(token, accountID),
		Marketplace: gqlclient.NewMarketplaceClient(token, accountID),

		Policies: policiesClient,
	}
}
