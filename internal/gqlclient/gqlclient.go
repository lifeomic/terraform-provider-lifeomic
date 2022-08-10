package gqlclient

//go:generate go run github.com/Khan/genqlient
//go:generate go run ./cmd/service-client-gen/... gqlclient appstore.go=AppStoreService:appstore.graphql marketplace.go=MarketplaceService:marketplace.graphql
