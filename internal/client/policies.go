package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

// PolicyService facilitates communication with the Policy-related endpoints of
// the PHC API.
// See: https://api.docs.lifeomic.com/#tag/Policy
type PolicyService interface {
	// List returns all policies.
	// See: https://api.docs.lifeomic.com/#tag/Policy/operation/list-policies
	List(context.Context, ListOptions) (PaginatedList[Policy], error)
	// Create creates a new policy.
	// See: https://api.docs.lifeomic.com/#tag/Policy/operation/create-policy
	Create(context.Context, *Policy) (*Policy, error)
	// Get gets a policy by name.
	// See: https://api.docs.lifeomic.com/#tag/Policy/operation/list-policies
	Get(context.Context, string) (*Policy, error)
	// Update updates an existing policy.
	// See: https://api.docs.lifeomic.com/#tag/Policy/operation/update-policy
	Update(context.Context, string, *Policy) (*Policy, error)
	// Delete deletes an existing policy.
	// See: https://api.docs.lifeomic.com/#tag/Policy/operation/delete-policy
	Delete(context.Context, string) error
}

// Policy represents an ABAC policy document, mapping operations to rules.
// See: https://devcenter.docs.lifeomic.com/development/abac-syntax
// Example Policy:
//
//		policy := client.Policy{
//	             Name: "my-policy",
//	             Policy: client.PolicyDocument{
//	                     Rules: client.Rules{{
//			      		"readData": client.Comparison{
//			              		"user.id": {
//			                		Comparison: client.ComparisonEquals,
//			                     		Value:      "johndoe",
//			                	},
//			        	},
//	             	}},
//		      	},
//	     }
type Policy struct {
	Name   string         `json:"name"`
	Policy PolicyDocument `json:"policy"`
}

// PolicyDocument represents an ABAC policy document.
// See: https://devcenter.docs.lifeomic.com/development/abac-syntax
type PolicyDocument struct {
	Rules PolicyRules `json:"rules"`
}

// PolicyRules maps operations to ABAC rules.
// See: https://devcenter.docs.lifeomic.com/development/abac-syntax#rules
// Example with staitc rule:
//
//	rules := PolicyRules{"readData": StaticRule(true)}
//
// Example with comparisons:
//
//	rules := PolicyRules{"readData": RuleMappings{
//	     	{"user.id": ValueComparison{Value: "johndoe", Type: ComparisonEquals}},
//	}}
type PolicyRules map[string]RuleExpression

func (r *PolicyRules) UnmarshalJSON(b []byte) error {
	var objectMap map[string]*json.RawMessage
	if err := json.Unmarshal(b, &objectMap); err != nil {
		return err
	}

	result := make(PolicyRules, len(objectMap))

	// The values of this map must either be bool or RuleMappings.
outer:
	for operation := range objectMap {
		var rawValue any
		if err := json.Unmarshal(*objectMap[operation], &rawValue); err != nil {
			return fmt.Errorf("could not parse rule for operation %q: %w", operation, err)
		}

		// Infer rule type from naive JSON type system.
		switch v := rawValue.(type) {
		case bool:
			// Upcast bool value
			result[operation] = StaticRule(v)
			continue outer
		case []any:
			// We can attempt to parse this is as RuleMappings.
			break
		default:
			return fmt.Errorf("unexpected primitive type for operation %q rule: %T", operation, v)
		}

		var rules RuleMappings
		if err := json.Unmarshal(*objectMap[operation], &rules); err != nil {
			return fmt.Errorf("could not parse rule mappings for operation %q: %w", operation, err)
		}
		result[operation] = rules
	}

	*r = result
	return nil
}

// RuleExpression represents a generic rule setting for an operation in ABAC.
// It's a wrapper around the polymorphic JSON which can either express a
// boolean value (StaticRule) or a list of mappings of attributes to
// comparisons (RuleMappings).
type RuleExpression interface {
	ruleExpression()
}

// StaticRule represents an operation rule which is either fully permissive or
// disabled. Use of this is heavily discouraged.
// See: https://devcenter.docs.lifeomic.com/development/abac-syntax#rules
type StaticRule bool

func (StaticRule) ruleExpression() {}

// RuleMappings is a slice of type RuleMaps.
type RuleMappings []RuleMap

func (RuleMappings) ruleExpression() {}

// RuleMap maps an attribute to an ABAC comparison.
// See: https://devcenter.docs.lifeomic.com/development/abac-syntax#rules
type RuleMap map[string]Comparison

// GetComparison attempts to get the single attribute/subject to Comparison
// mapping that an ABAC rule should consist of. Returns false for malformed
// RuleMaps.
func (r RuleMap) GetComparison() (subject string, comparison Comparison, ok bool) {
	if len(r) == 1 {
		for subject, comparison := range r {
			return subject, comparison, true
		}
	}
	return "", nil, false
}

func (r *RuleMap) UnmarshalJSON(b []byte) error {
	var ruleMap map[string]*json.RawMessage
	if err := json.Unmarshal(b, &ruleMap); err != nil {
		return err
	}

	size := len(ruleMap)
	if size != 1 {
		return fmt.Errorf("failed to parse RuleMap: should have exactly one entry, has %d", size)
	}

	// This is ugly but far less expensive than
	// reflect.ValueOf(map).MapKeys().
	var attribute string
	for key := range ruleMap {
		attribute = key
	}

	var comparisonMap map[string]any
	if err := json.Unmarshal(*ruleMap[attribute], &comparisonMap); err != nil {
		return err
	}

	var comparison Comparison
	switch {
	case mapHasKey(comparisonMap, "value"):
		if _, ok := comparisonMap["value"].(string); ok {
			comparison = &ValueComparison{}
			break
		}
		comparison = &MultivalueComparison{}

	case mapHasKey(comparisonMap, "target"):
		comparison = &TargetComparison{}

	default:
		return errors.New("malformed comparison object")
	}

	if err := json.Unmarshal(*ruleMap[attribute], comparison); err != nil {
		return fmt.Errorf("failed to parse comparison: %w", err)
	}

	*r = RuleMap{attribute: comparison}
	return nil
}

// Comparison represents a generic ABAC comparison. It's a wrapper around the
// polymorphic JSON which can either express a comparison between an attribute
// and another attribute (TargetComparison), a single value (ValueComparison),
// or an array of values (MultivalueComparison).
// See: https://devcenter.docs.lifeomic.com/development/abac-syntax#comparisons
type Comparison interface {
	GetComparisonType() ComparisonType
}

// ValueComparison represents an ABAC comparison between an attribute and some
// value.
// See: https://devcenter.docs.lifeomic.com/development/abac-syntax#comparisons
type ValueComparison struct {
	Comparison ComparisonType `json:"comparison"`
	Value      string         `json:"value"`
}

func (c ValueComparison) GetComparisonType() ComparisonType { return c.Comparison }

// MultivalueComparison represents an ABAC comparison between an attirbute and
// some values.
// See: https://devcenter.docs.lifeomic.com/development/abac-syntax#comparisons
type MultivalueComparison struct {
	Comparison ComparisonType `json:"comparison"`
	Values     []string       `json:"value"`
}

func (c MultivalueComparison) GetComparisonType() ComparisonType { return c.Comparison }

// TargetComparison represents an ABAC comparison against some attribute.
type TargetComparison struct {
	Comparison ComparisonType `json:"comparison"`
	Target     string         `json:"target"`
}

func (c TargetComparison) GetComparisonType() ComparisonType { return c.Comparison }

// A ComparisonType represents an ABAC comparison type.
// See: https://devcenter.docs.lifeomic.com/development/abac-syntax#supported-comparisons
type ComparisonType string

// supported comparison constants
const (
	ComparisonEquals      ComparisonType = "equals"
	ComparisonNotEquals   ComparisonType = "notEquals"
	ComparisonIncludes    ComparisonType = "includes"
	ComparisonNotIncludes ComparisonType = "notIncludes"
	ComparisonIn          ComparisonType = "in"
	ComparisonNotIn       ComparisonType = "notIn"
	ComparisonExists      ComparisonType = "exists"
	ComparisonSuperset    ComparisonType = "superset"
	ComparisonSubset      ComparisonType = "subset"
	ComparisonStartsWith  ComparisonType = "startsWith"
	ComparisonPrefixOf    ComparisonType = "prefixOf"
	ComparisonEndsWith    ComparisonType = "endsWith"
	ComparisonSuffixOf    ComparisonType = "suffixOf"
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

func (s *policyService) Get(ctx context.Context, id string) (*Policy, error) {
	id = url.PathEscape(id)
	res, err := checkResponse(s.Request(ctx).SetResult(&Policy{}).Get("/policies/" + id))
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

func mapHasKey[K comparable, V any](m map[K]V, key K) bool {
	_, ok := m[key]
	return ok
}
