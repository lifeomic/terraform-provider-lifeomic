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
	List(context.Context, ListOptions) (PaginatedList[Account], error)
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

// AccountList contains a slice of Accounts and pagination fields.
type accountList struct {
	ListResponse

	Accounts []Account `json:"accounts"`

	listOptions    ListOptions    `json:"-"`
	accountService AccountService `json:"-"`
}

func (l *accountList) GetNextPage(ctx context.Context) (PaginatedList[Account], error) {
	if !l.HasNextPage() {
		return nil, ErrNoNextPage
	}

	options := l.listOptions
	options.NextPageToken = l.GetNextPageToken()
	return l.accountService.List(ctx, options)
}

func (l *accountList) Items() []Account { return l.Accounts }

func (s *accountService) List(ctx context.Context, options ListOptions) (PaginatedList[Account], error) {
	req := s.Request(ctx).SetResult(&accountList{})
	req.Header.Del(accountHeader)

	endpoint, err := buildQueryURL("/accounts", &options)
	if err != nil {
		return nil, err
	}

	res, err := checkResponse(req.Get(endpoint))
	if err != nil {
		return nil, err
	}

	accountList := res.Result().(*accountList)
	accountList.accountService = s
	accountList.listOptions = options
	return accountList, nil
}
