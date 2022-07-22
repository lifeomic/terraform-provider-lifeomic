package main

import (
	"context"

	"github.com/lifeomic/terraform-provider-phc/internal/client"
)

func main() {
	ctx := context.Background()
	cli := client.New(client.Config{
		Account: "tfprovidertest",
	})

	policyClient := cli.Policies()

	// Create new policy.
	rules := client.PolicyRules{
		"readMaskedData": {{
			"user.id": client.Comparison{
				Comparison: client.ComparisonEquals,
				Value:      "johndoe",
			},
		}},
	}
	newPolicy, err := policyClient.Create(ctx, &client.Policy{
		Name:   "foo",
		Policy: client.PolicyDocument{Rules: rules},
	})
	if err != nil {
		panic(err)
	}

	// Rename the policy.
	updatedPolicy, err := policyClient.Update(ctx, newPolicy.Name, &client.Policy{
		Name: "bar",
		Policy: client.PolicyDocument{
			Rules: rules,
		},
	})
	if err != nil {
		panic(err)
	}

	// Delete the policy.
	err = policyClient.Delete(ctx, updatedPolicy.Name)
	if err != nil {
		panic(err)
	}
}
