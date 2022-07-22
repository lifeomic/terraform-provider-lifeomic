package main

import (
	"context"
	"fmt"

	"github.com/lifeomic/terraform-provider-phc/internal/client"
)

func main() {
	ctx := context.Background()
	cli := client.New(client.Config{
		Account: "tfprovidertest",
	})

	// List accounts
	accountList, err := cli.Accounts().List(ctx, client.ListOptions{
		PageSize: 1,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("accountList: %+v\n", accountList)

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

	// List all policies with pagination.
	policyList, err := policyClient.List(ctx, client.ListOptions{
		PageSize: 2,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("policies: %+v", policyList.Items())

	for policyList.HasNextPage() {
		policyList, err = policyList.GetNextPage(ctx)
		if err != nil {
			panic(err)
		}

		fmt.Printf("paged policies: %+v\n", policyList)
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
