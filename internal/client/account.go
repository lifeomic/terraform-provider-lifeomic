package client

import (
	"context"
)

// AccountService facilitates communication with the Account-related endpoints
// of the PHC API.
// See: https://docs.us.lifeomic.com/api/#lifeomic-core-api-accounts
type AccountService interface {
	// Returns a list of accounts that the user has access to.
	// See: https://docs.us.lifeomic.com/api/#list-all-accounts
	List(context.Context) ([]Account, error)
}

// AccountType represents the type of an Account.
type AccountType string

// AccountType constants
const (
	AccountTypeFree       AccountType = "free"
	AccountTypePaid       AccountType = "paid"
	AccountTypeEnterprise AccountType = "enterprise"
)

// Account represents a PHC Account.
type Account struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

type accountService struct {
	*Client
}

// accountService implements AccountService.
var _ AccountService = &accountService{}

type accountListResponse struct {
	Accounts []Account `json:"accounts"`
}

func (s *accountService) List(ctx context.Context) ([]Account, error) {
	req := s.Request(ctx).SetResult(&accountListResponse{})
	req.Header.Del(accountHeader)

	res, err := checkResponse(req.Get("/accounts"))
	if err != nil {
		return nil, err
	}

	return res.Result().(*accountListResponse).Accounts, nil
}
