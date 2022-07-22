package client

import (
	"context"
	"errors"
	"net/url"

	"github.com/dyninc/qstring"
)

// pagination errors
var (
	ErrNoNextPage = errors.New("no next page to fetch")
)

// PaginatedList represents a list of a resource T which can be paginated.
type PaginatedList[T any] interface {
	HasNextPage() bool
	GetNextPage(context.Context) (PaginatedList[T], error)
	GetNextPageToken() string

	Items() []T
}

// ListOptions represent parameters for generic List requests.
type ListOptions struct {
	NextPageToken string `qstring:"nextPageToken,omitempty"`
	PageSize      int    `qstring:"pageSize,omitempty"`
}

// ListLinks include links related to the resource list.
type ListLinks struct {
	Self string  `json:"self"`
	Next *string `json:"next"`
}

// ListResponse represents the base object for generic list responses.
type ListResponse struct {
	Links *ListLinks `json:"links"`
}

func (r *ListResponse) HasNextPage() bool {
	return r.Links != nil && r.Links.Next != nil
}

// GetNextPageToken returns a nextPageToken or an empty string.
func (r *ListResponse) GetNextPageToken() string {
	if !r.HasNextPage() {
		return ""
	}

	url, err := url.Parse(*r.Links.Next)
	if err != nil {
		return ""
	}
	return url.Query().Get("nextPageToken")
}

// buildQueryURL formats an endpoint with query parameters.
func buildQueryURL[T any](endpoint string, params *T) (string, error) {
	query, err := qstring.MarshalString(params)
	if err != nil {
		return "", err
	}

	if query == "" {
		return endpoint, nil
	}
	return endpoint + "?" + query, nil
}
