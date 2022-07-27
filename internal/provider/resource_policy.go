package provider

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/lifeomic/terraform-provider-phc/internal/client"
)

func resourcePolicy() *schema.Resource {
	return &schema.Resource{
		Description: "`phc_policy` manages an individual access control policy.",

		ReadContext:   resourcePolicyRead,
		CreateContext: resourcePolicyCreate,
		UpdateContext: resourcePolicyUpdate,
		DeleteContext: resourcePolicyDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the policy.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"rule": {
				Description: "A rule to attach to the policy.",
				Type:        schema.TypeList,
				Elem:        policyRuleResource(),
				Required:    true,

				// TODO: implement SchemaDiffSuppressFunc to
				// compare the spec of rules against the state of
				// the world without order mattering.
				RequiredWith: []string{"name"},
			},
		},
	}
}

func policyRuleResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"operation": {
				Description: "The operation this policy rule governs.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"comparison": {
				Description: "The ABAC attribute comparison of this policy rule.",
				Required:    true,
				Type:        schema.TypeList,
				Elem:        policyRuleComparisonResource(),
			},
		},
	}
}

func policyRuleComparisonResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description: "The type of ABAC comparison.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					string(client.ComparisonEndsWith),
					string(client.ComparisonEquals),
					string(client.ComparisonExists),
					string(client.ComparisonIn),
					string(client.ComparisonIncludes),
					string(client.ComparisonNotEquals),
					string(client.ComparisonNotIn),
					string(client.ComparisonStartsWith),
					string(client.ComparisonSubset),
					string(client.ComparisonSuperset),
				}, false),
			},
			"subject": {
				Description: "The subject of this comparison.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"value": {
				Description: "The value to compare the subject to.",
				Type:        schema.TypeString,
				Optional:    true,

				ConflictsWith: []string{"values", "target"},
			},
			"values": {
				Description: "The values to compare the subject to.",
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,

				ConflictsWith: []string{"value", "target"},
			},
			"target": {
				Description: "The attribute to compare the subject to.",
				Type:        schema.TypeString,
				Optional:    true,

				ConflictsWith: []string{"value", "values"},
			},
		},
	}
}

// expandPolicyDocument expands the rules in the shape of the phc_policy.rule's
// schema into the structs used by the go-client so changes can be pushed up to
// the API accordingly.
func expandPolicyDocument(rules []any) (*client.PolicyDocument, error) {
	policy := client.PolicyDocument{Rules: make(client.PolicyRules)}
	for i, rule := range rules {
		ruleSpec := rule.(map[string]any)
		operation := ruleSpec["operation"].(string)

		// Expand and accumulate each comparison for all operations
		// from the spec.
		for j, comparision := range ruleSpec["comparison"].([]any) {
			comparisonSpec := comparision.(map[string]any)
			subject := comparisonSpec["subject"].(string)
			comparisonType := client.ComparisonType(comparisonSpec["type"].(string))

			var comparison client.Comparison

			// Build attribute-comparison mapping for the current rule.
			if value, ok := comparisonSpec["value"]; ok {
				comparison = &client.ValueComparison{
					Comparison: comparisonType,
					Value:      value.(string),
				}
			} else if target, ok := comparisonSpec["target"]; ok {
				comparison = &client.TargetComparison{
					Comparison: comparisonType,
					Target:     target.(string),
				}
			} else if values, ok := comparisonSpec["values"]; ok {
				// Typecast slice of interfaces to strings.
				values := values.([]any)
				stringValues := make([]string, len(values))
				for i := range values {
					stringValues[i] = values[i].(string)
				}

				comparison = &client.MultivalueComparison{
					Comparison: comparisonType,
					Values:     stringValues,
				}
			} else {
				// If we've gotten to this point, no valid
				// client.Comparison struct can be built. The
				// resource is missing some pseudo-required
				// fields.
				// TODO: use diag.Diagnostic with AttributePath.
				return nil, fmt.Errorf("either value, values, or target must be set on .rule.%d.comparison.%d", i, j)
			}

			ruleMappings := policy.Rules[operation].(client.RuleMappings)
			policy.Rules[operation] = append(ruleMappings, map[string]client.Comparison{
				subject: comparison,
			})
		}
	}

	return &policy, nil
}

// flattenPolicyDocument flattens the rules for each operation specified in the
// given client.PolicyDocument into the schema of phc_policy.rule objects.
//
// If the schema of the response is malformed or cannot be represented in the
// Terraform schema, a warning will be emitted, but not an error -- all schema
// differences will be otherwise ignored and resolved after the next apply is
// executed.
func flattenPolicyDocument(document client.PolicyDocument) (flattend []map[string]any, diagnostics diag.Diagnostics) {
	flattend = make([]map[string]any, len(document.Rules))
	diagnostics = make(diag.Diagnostics, 0)

	for i, operation := range reflect.ValueOf(document.Rules).MapKeys() {
		rulesValue := document.Rules[operation.String()]

		// Ensure this operation's rules are valid for Terraform.
		if value, ok := rulesValue.(client.StaticRule); ok {
			// StaticRules cannot be expressed in Terraform because
			// they break the shape of the schema and are bad
			// practice. We will interpret this as an empty slice
			// and next apply will overwrite this rule with either
			// nothing or a valid set of comparisons.
			return make([]map[string]any, 0), []diag.Diagnostic{{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("found a static rule of %t for operation %q; this is unsupported and will be corrected on next apply", value, operation),
				Detail:   "This change had to have been made outside of Terraform. Setting an operation to a boolean is unsupported as it's bad practice.",
			}}
		}

		// If we got here, rules has to be a slice of RuleMaps.
		rules := rulesValue.(client.RuleMappings)
		comparisons := make([]map[string]any, len(rules))

		// Extract the comparisons from each rule for the current
		// operation.
		for j, rule := range rules {
			subject, comparison, ok := rule.GetComparison()
			if !ok {
				// This *should* never happen as
				// RuleMappings.UnmarshalJSON will throw an error
				// if len != 1 and will be propagated on the call
				// to PolicyService.Get earlier in the scope of
				// the current Terraform operation.
				diagnostics = append(diagnostics, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  fmt.Sprintf("found a malformed comparison [%d] for operation %q; will be corrected on next apply", j, operation),
				})
				comparisons[j] = make(map[string]any)
				continue
			}

			comparisonMap := map[string]any{
				"subject": subject,
				"type":    comparison.GetComparisonType(),
			}

			// This switch statement is exhaustive of all
			// possible client.Comparison structs.
			switch c := comparison.(type) {
			case *client.ValueComparison:
				comparisonMap["value"] = c.Value
			case *client.MultivalueComparison:
				comparisonMap["values"] = c.Values
			case *client.TargetComparison:
				comparisonMap["target"] = c.Target
			}
			comparisons[j] = comparisonMap
		}

		flattend[i] = map[string]any{
			"operation":  operation.String(),
			"comparison": comparisons,
		}
	}
	return
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyService := meta.(*providerMeta).Client.Policies()

	name := d.Id()
	policy, err := policyService.Get(ctx, name)
	if err != nil {
		// handle 404 here as deleting the id
		return diag.Errorf("failed to get the policy %q: %s", name, err)
	}

	fmt.Printf("[INFO] rules = %+v", policy.Policy.Rules)

	d.Set("name", policy.Name)

	rules, diagnositcs := flattenPolicyDocument(policy.Policy)
	if err := d.Set("rule", rules); err != nil {
		return append(diagnositcs, diag.FromErr(err)...)
	}

	return diagnositcs
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyService := meta.(*providerMeta).Client.Policies()

	name := d.Get("name").(string)
	policyDocument, err := expandPolicyDocument(d.Get("rule").([]any))
	if err != nil {
		return diag.FromErr(err)
	}

	policy, err := policyService.Create(ctx, &client.Policy{
		Name:   name,
		Policy: *policyDocument,
	})
	if err != nil {
		return diag.Errorf("failed to create policy: %s", err)
	}

	d.SetId(policy.Name)
	d.Set("name", policy.Name)

	rules, diagnositcs := flattenPolicyDocument(policy.Policy)
	if err := d.Set("rule", rules); err != nil {
		return append(diagnositcs, diag.FromErr(err)...)
	}

	return nil
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyService := meta.(*providerMeta).Client.Policies()

	name := d.Id()
	policyDocument, err := expandPolicyDocument(d.Get("rule").([]any))
	if err != nil {
		return diag.FromErr(err)
	}

	policy, err := policyService.Update(ctx, name, &client.Policy{
		Name:   name,
		Policy: *policyDocument,
	})
	if err != nil {
		return diag.Errorf("failed to update policy %q: %s", name, err)
	}

	d.Set("name", policy.Name)

	rules, diagnositcs := flattenPolicyDocument(policy.Policy)
	if err := d.Set("rule", rules); err != nil {
		return append(diagnositcs, diag.FromErr(err)...)
	}

	return nil
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyService := meta.(*providerMeta).Client.Policies()

	name := d.Get("name").(string)
	if err := policyService.Delete(ctx, name); err != nil {
		return diag.Errorf("failed to delete policy %s: %s", name, err)
	}
	d.SetId("")
	return nil
}
