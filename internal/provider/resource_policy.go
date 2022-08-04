package provider

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/lifeomic/terraform-provider-phc/internal/client"
)

const (
	policyDocsURL                     = "https://phc.docs.lifeomic.com/user-guides/access-control#privileges-and-permissions"
	policyRuleDocsURL                 = "https://phc.docs.lifeomic.com/development/abac-syntax#rules"
	policyComparisonDocsURL           = "https://phc.docs.lifeomic.com/development/abac-syntax#comparisons"
	policyOperationDocsURL            = "https://phc.docs.lifeomic.com/development/abac-syntax#operations"
	policyAttributeDocsURL            = "https://phc.docs.lifeomic.com/development/abac-syntax#attributes"
	policySupportedComparisonsDocsURL = "https://phc.docs.lifeomic.com/development/abac-syntax#supported-comparisons"
)

// policy represents the state of a phc_policy resource.
type policy struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Rule []policyRule `tfsdk:"rule"`
}

// policyRule represents the state of a phc_policy resource's rule block.
type policyRule struct {
	Operation  types.String           `tfsdk:"operation"`
	Comparison []policyRuleComparison `tfsdk:"comparison"`
}

// policyRuleComparison represents the state of a phc_policy resource's
// comparison block.
type policyRuleComparison struct {
	Type    types.String `tfsdk:"type"`
	Subject types.String `tfsdk:"subject"`
	Value   *string      `tfsdk:"value"`
	Values  *[]string    `tfsdk:"values"`
	Target  *string      `tfsdk:"target"`
}

// policyRuleWalkFunc is called with the a policyRule struct and it's index.
type policyRuleWalkFunc func(index int, rule *policyRule)

// policyResource implements tfsdk.
type policyResource struct {
	client client.Interface
}

// policyResource implements tfsdk.ResourceType
type policyResourceType struct {
}

func (policyResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: fmt.Sprintf("`phc_policy` manages an [Attribute Based Access Control (ABAC) policy](%s).", policyDocsURL),
		Attributes: map[string]tfsdk.Attribute{
			"name": {
				Required:    true,
				Type:        types.StringType,
				Description: "The unique name of this ABAC policy.",
			},
			"id": {
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "The ID of this ABAC policy resource.",
			},
		},
		Blocks: map[string]tfsdk.Block{
			"rule": {
				Description: fmt.Sprintf("An ABAC [rule](%s) containing comparisons to be evaluated for the given operation.", policyRuleDocsURL),
				NestingMode: tfsdk.BlockNestingModeList,
				Validators: []tfsdk.AttributeValidator{
					&policyRulesValidator{},
				},
				MinItems: 1,
				Attributes: map[string]tfsdk.Attribute{
					"operation": {
						Type:        types.StringType,
						Required:    true,
						Description: fmt.Sprintf("The [operation](%s) this ABAC rule governs.", policyOperationDocsURL),
					},
				},
				Blocks: map[string]tfsdk.Block{
					"comparison": {
						Description: "An ABAC comparison. Exactly one of `value`, `values`, or `target` should be set.",
						MinItems:    1,
						NestingMode: tfsdk.BlockNestingModeList,
						Validators: []tfsdk.AttributeValidator{
							&policyRuleComparisonValidator{},
						},
						Attributes: map[string]tfsdk.Attribute{
							"type": {
								Type:        types.StringType,
								Required:    true,
								Description: fmt.Sprintf("The [type](%s) of ABAC comparison.", policySupportedComparisonsDocsURL),
							},
							"subject": {
								Type:        types.StringType,
								Required:    true,
								Description: fmt.Sprintf("The subject is the [attribute](%s) used in this ABAC comparison.", policyAttributeDocsURL),
							},
							"values": {
								Type:        types.ListType{ElemType: types.StringType},
								Optional:    true,
								Description: "The values to use in this ABAC comparison.",
							},
							"value": {
								Type:        types.StringType,
								Optional:    true,
								Description: "The value to use in this ABAC comparison.",
							},
							"target": {
								Type:        types.StringType,
								Optional:    true,
								Description: "The target to use in this ABAC comparison.",
							},
						},
					},
				},
			},
		},
	}, nil
}

func (policyResourceType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	pr, ok := p.(*provider)
	if !ok {
		return nil, errorConvertingProvider(p)
	}

	return &policyResource{
		client: pr.client,
	}, nil
}

// walkPolicyRuleList attempts to cast the given tfsdk.List as a list of
// policyRule structs and visits each element, calling the policyRuleWalkFunc.
//
// If the list is invalid or has duplicate operations, an error will be
// returned.
func walkPolicyRuleList(ctx context.Context, basePath path.Path, rules []policyRule, walkFunc policyRuleWalkFunc) (diags diag.Diagnostics) {
	// Map rule operations to their respective indicies in the rule list.
	// If there are multiple occurances of an operation, tell the user to
	// consolidate their rules.
	// All errors with the rule configuration will be accumulated before
	// returning.
	seen := make(map[string]int, len(rules))

	for i, rule := range rules {
		operation := rule.Operation.Value

		// Ensure we haven't seen this operation in a previous rule
		// block.
		if location, ok := seen[operation]; ok {
			// This is a duplicate operation.
			diags.AddAttributeError(basePath.AtListIndex(i),
				fmt.Sprintf("Duplicate occurance of rule for operation %q", operation),
				fmt.Sprintf("Consolidate %q rules with block at index %d", operation, location))
			continue
		}

		// Record this index for the given operation.
		seen[rule.Operation.Value] = i

		// Visit this element.
		walkFunc(i, &rule)
	}

	return
}

// ToPolicyObject converts a policy resource's state to a client.Policy struct.
func (p *policy) ToPolicyObject(ctx context.Context) (policy *client.Policy, diags diag.Diagnostics) {
	policy = new(client.Policy)
	policy.Name = p.Name.Value

	// Build client.RuleMappings from resource struct.
	rules := make(client.PolicyRules, len(p.Rule))

	walkPolicyRuleList(ctx, path.Root("rule"), p.Rule, func(index int, rule *policyRule) {
		operation := rule.Operation.Value
		ruleMappings := make(client.RuleMappings, len(rule.Comparison))

		for i, comparisonSpec := range rule.Comparison {
			comparisonType := client.ComparisonType(comparisonSpec.Type.Value)
			subject := comparisonSpec.Subject.Value

			// Determine the type of the comparison and build the
			// appropriate struct to add to the policy.
			var comparison client.Comparison

			switch {
			case comparisonSpec.Value != nil:
				comparison = client.ValueComparison{
					Comparison: comparisonType,
					Value:      *comparisonSpec.Value,
				}

			case comparisonSpec.Target != nil:
				comparison = client.TargetComparison{
					Comparison: comparisonType,
					Target:     *comparisonSpec.Target,
				}

			case comparisonSpec.Values != nil && len(*comparisonSpec.Values) != 0:
				comparison = client.MultivalueComparison{
					Comparison: comparisonType,
					Values:     *comparisonSpec.Values,
				}
			}

			ruleMappings[i] = client.RuleMap{
				subject: comparison,
			}
		}

		// Add rules for operation.
		rules[operation] = ruleMappings
	})

	policy.Policy.Rules = rules
	return
}

type emptyValidatorDescriptions struct{}

func (emptyValidatorDescriptions) MarkdownDescription(_ context.Context) string {
	return ""
}

func (emptyValidatorDescriptions) Description(_ context.Context) string {
	return ""
}

// policyRulesValidator is a tfsdk.AttributeValidator for the phc_policy.rule
// blocks.
type policyRulesValidator struct {
	emptyValidatorDescriptions
}

// Validate ensures that the phc_policy.rule blocks are not duplicates. An
// operation should only be referenced in a single block.
func (v *policyRulesValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	if req.AttributeConfig.IsUnknown() {
		// It's okay if the values are unknown at plan time. It's means
		// the user has referenced other resources/datasources/dynamic
		// values inside a rule block.
		//
		// Values will be known at apply-time and any validation
		// errors will be surfaced then before things are actually
		// changed.
		return
	}

	var list []policyRule
	if diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &list); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(walkPolicyRuleList(ctx, req.AttributePath, list, func(_ int, _ *policyRule) {
		// Noop. Nothing to process, walkPolicyRuleList will validate
		// the rule blocks.
	})...)
}

type policyRuleComparisonValidator struct {
	emptyValidatorDescriptions
}

// Validate enures that the phc_policy.rule[*].comparison blocks only specify
// one of the target, value, or values fields.
func (v *policyRuleComparisonValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	if req.AttributeConfig.IsUnknown() {
		return
	}

	var comparisons []policyRuleComparison
	resp.Diagnostics.Append(tfsdk.ValueAs(ctx, req.AttributeConfig, &comparisons)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for i, comparison := range comparisons {
		set := make([]string, 0, 3)
		if comparison.Target != nil {
			set = append(set, "target")
		}
		if comparison.Value != nil {
			set = append(set, "value")
		}
		if comparison.Values != nil {
			set = append(set, "values")
		}

		if len(set) != 1 {
			resp.Diagnostics.AddAttributeError(req.AttributePath.AtListIndex(i),
				"Exactly one of value, values, or target must be set",
				fmt.Sprintf("Unset one of %s", set))
		}
	}
}

func (r policyResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	tflog.Info(ctx, "Creating Policy resource")

	// Get plan values.
	var plan policy
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map Terraform plan to *client.Policy object.
	p, diags := plan.ToPolicyObject(ctx)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Create the policy.
	p, err := r.client.Policies().Create(ctx, p)
	if err != nil {
		resp.Diagnostics.AddError("failed to create policy", err.Error())
		return
	}

	tflog.Info(ctx, "Created new Policy", map[string]any{"policy": p})
	resp.Diagnostics.Append(setPolicyState(ctx, &plan, &resp.State, p)...)
}

func (r policyResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	tflog.Info(ctx, "Reading Policy resource")

	// Get current state.
	var state policy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the underlying Policy object.
	p, err := r.client.Policies().Get(ctx, state.Name.Value)
	if err != nil {
		resp.Diagnostics.AddError("failed to get policy", err.Error())
		return
	}

	tflog.Info(ctx, "Got existing Policy", map[string]any{"policy": p})
	resp.Diagnostics.Append(setPolicyState(ctx, &state, &resp.State, p)...)
}

func (r policyResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	tflog.Info(ctx, "Updating Policy resource")

	// Get plan values.
	var plan policy
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state.
	var state policy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p, diags := plan.ToPolicyObject(ctx)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	p, err := r.client.Policies().Update(ctx, state.Name.Value, p)
	if err != nil {
		resp.Diagnostics.AddError("failed to create policy", err.Error())
		return
	}

	tflog.Info(ctx, "Updated existing Policy", map[string]any{"policy": p})
	resp.Diagnostics.Append(setPolicyState(ctx, &plan, &resp.State, p)...)
}

func (r policyResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	tflog.Info(ctx, "Deleting Policy resource")

	// Get current state.
	var state policy
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Policies().Delete(ctx, state.Name.Value); err != nil {
		resp.Diagnostics.AddError("failed to delete Policy", err.Error())
	}

	resp.State.RemoveResource(ctx)
	tflog.Info(ctx, "Deleted existing Policy", map[string]any{"name": state.Name})
}

func setPolicyState(ctx context.Context, config *policy, state *tfsdk.State, p *client.Policy) (diags diag.Diagnostics) {
	rules := make([]policyRule, 0, len(p.Policy.Rules))

	// Set the rules in the same order as they are declared in the plan
	// generated from the user's config to avoid messy diffs.
	// UnmarshalJSON does not order maps as they are parsed.
	walkPolicyRuleList(ctx, path.Root("rule"), config.Rule, func(index int, ruleSpec *policyRule) {
		operation := ruleSpec.Operation.Value
		ruleElem := policyRule{Operation: types.String{Value: operation}}

		// Ensure this operation's rules are valid for Terraform
		if _, ok := p.Policy.Rules[operation].(client.StaticRule); ok {
			// StaticRules cannot be expressed in Terraform because
			// they break the shape of the schema and are bad
			// practice. We will interpret this as an empty slice
			// and next apply will overwrite this rule with either
			// nothing or a valid set of comparisons.
			diags.AddWarning(fmt.Sprintf("Detected incompatible rule for operation %s", operation),
				"Using a static boolean value for rules is bad practice and not supported. "+
					"The changes happened outside of Terraform")
			return
		}

		ruleMappings := p.Policy.Rules[operation].(client.RuleMappings)

		// Build comparison objects.
		comparisons := make([]policyRuleComparison, len(ruleMappings))
		for i, ruleMap := range ruleMappings {
			subject, comparison, ok := ruleMap.GetComparison()
			if !ok {
				continue
			}

			comparisonElem := policyRuleComparison{
				Subject: types.String{Value: subject},
				Type:    types.String{Value: string(comparison.GetComparisonType())},
			}

			switch c := comparison.(type) {
			case *client.ValueComparison:
				comparisonElem.Value = &c.Value
			case *client.MultivalueComparison:
				comparisonElem.Values = &c.Values
			case *client.TargetComparison:
				comparisonElem.Target = &c.Target
			}
			comparisons[i] = comparisonElem
		}

		ruleElem.Comparison = comparisons
		rules = append(rules, ruleElem)

		// Removed processed rule block from the policy object to keep
		// track of remaining rules.
		delete(p.Policy.Rules, operation)
	})

	// Ensure all rules have been mapped to the Terraform state. If there
	// are still values left over after walking the rule blocks, there must
	// have been some changes outside of Terraform.
	if len(p.Policy.Rules) != 0 {
		keys := reflect.ValueOf(p.Policy.Rules).MapKeys()
		diags.AddWarning("Detected state drift",
			fmt.Sprintf("Found rules for %q operations that don't exist in config. These changes happened outside of Terraform", keys))
	}

	if diags.HasError() {
		return
	}

	diags.Append(state.Set(ctx, policy{
		ID:   types.String{Value: p.Name},
		Name: types.String{Value: p.Name},
		Rule: rules,
	})...)
	return
}
