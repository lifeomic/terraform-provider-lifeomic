package client

import (
	"context"
	"net/url"
)

// PolicyService facilitates communication with the Policy-related endpoints of
// the PHC API.
// See: https://docs.us.lifeomic.com/api/#lifeomic-core-api-policy
type PolicyService interface {
	// List returns all policies.
	// See: https://docs.us.lifeomic.com/api/#list-policies
	List(context.Context, ListOptions) (PaginatedList[Policy], error)
	// Create creates a new policy.
	// See: https://docs.us.lifeomic.com/api/#create-policy
	Create(context.Context, *Policy) (*Policy, error)
	// Update updates an existing policy.
	// See: https://docs.us.lifeomic.com/api/#update-a-policy
	Update(context.Context, string, *Policy) (*Policy, error)
	// Delete deletes an existing policy.
	// See: https://docs.us.lifeomic.com/api/#delete-a-policy
	Delete(context.Context, string) error
}

// Policy represents an ABAC policy document, mapping operations to rules.
// See: https://phc.docs.lifeomic.com/development/abac-syntax
// Example Policy:
// 	policy := client.Policy{
//              Name: "my-policy",
//              Policy: client.PolicyDocument{
//                      Rules: client.Rules{{
//		      		"readData": client.Comparison{
//		              		"user.id": {
//		                		Comparison: client.ComparisonEquals,
//		                     		Value:      "johndoe",
//		                	},
//		        	},
//              	}},
//	      	},
//      }
type Policy struct {
	Name   string         `json:"name"`
	Policy PolicyDocument `json:"policy"`
}

// PolicyDocument represents an ABAC policy document.
// See: https://phc.docs.lifeomic.com/development/abac-syntax
type PolicyDocument struct {
	Rules PolicyRules `json:"rules"`
}

// PolicyRules maps operations to ABAC rules.
// See: https://phc.docs.lifeomic.com/development/abac-syntax#rules
type PolicyRules map[string][]Rule

// Rule maps contextual values to conditions using ABAC comparisons.
// See: https://phc.docs.lifeomic.com/development/abac-syntax#rules
type Rule map[string]Comparison

// Comparison represents an ABAC comparison.
// See: https://phc.docs.lifeomic.com/development/abac-syntax#rules
type Comparison struct {
	Comparison ComparisonType `json:"comparison,omitempty"`
	Value      string         `json:"value,omitempty"`
	Target     string         `json:"target,omitempty"`
}

// A ComparisonType represents an ABAC comparison type.
// See: https://phc.docs.lifeomic.com/development/abac-syntax#supported-comparisons
type ComparisonType string

// supported comparison constants
const (
	ComparisonEquals     ComparisonType = "equals"
	ComparisonNotEquals  ComparisonType = "notEquals"
	ComparisonIncludes   ComparisonType = "includes"
	ComparisonIn         ComparisonType = "in"
	ComparisonNotIn      ComparisonType = "notIn"
	ComparisonExists     ComparisonType = "exists"
	ComparisonSuperset   ComparisonType = "superset"
	ComparisonSubset     ComparisonType = "subset"
	ComparisonStartsWith ComparisonType = "startsWith"
	ComparisonEndsWith   ComparisonType = "endsWith"
)

type policyService struct {
	*Client
}

// policyService implements PolicyService.
var _ PolicyService = &policyService{}

type policyList struct {
	ListResponse

	Policies []Policy `json:"items"`

	listOptions   ListOptions   `json:"-"`
	policyService PolicyService `json:"-"`
}

func (l *policyList) GetNextPage(ctx context.Context) (PaginatedList[Policy], error) {
	if !l.HasNextPage() {
		return nil, ErrNoNextPage
	}

	options := l.listOptions
	options.NextPageToken = l.GetNextPageToken()
	return l.policyService.List(ctx, options)
}

func (l *policyList) Items() []Policy { return l.Policies }

func (s *policyService) List(ctx context.Context, options ListOptions) (PaginatedList[Policy], error) {
	endpoint, err := buildQueryURL("/policies", &options)
	if err != nil {
		return nil, err
	}

	res, err := checkResponse(s.Request(ctx).SetResult(&policyList{}).Get(endpoint))
	if err != nil {
		return nil, err
	}

	policyList := res.Result().(*policyList)
	policyList.policyService = s
	policyList.listOptions = options
	return policyList, nil
}

func (s *policyService) Create(ctx context.Context, policy *Policy) (*Policy, error) {
	res, err := checkResponse(s.Request(ctx).SetBody(policy).SetResult(&Policy{}).Post("/policies"))
	if err != nil {
		return nil, err
	}
	return res.Result().(*Policy), nil
}

func (s *policyService) Update(ctx context.Context, id string, policy *Policy) (*Policy, error) {
	id = url.PathEscape(id)
	res, err := checkResponse(s.Request(ctx).SetBody(policy).SetResult(&Policy{}).Put("/policies/" + id))
	if err != nil {
		return nil, err
	}
	return res.Result().(*Policy), nil
}

func (s *policyService) Delete(ctx context.Context, id string) error {
	id = url.PathEscape(id)
	_, err := checkResponse(s.Request(ctx).Delete("/policies/" + id))
	return err
}
