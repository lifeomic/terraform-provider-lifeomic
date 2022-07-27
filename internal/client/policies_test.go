package client

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRuleMap_UnmarshalJSON(t *testing.T) {
	for _, fixture := range []struct {
		name          string
		json          string
		expectedValue RuleMap
		expectedErr   string
	}{
		{
			name:          "should parse target comparison",
			json:          `{"user.patients": {"comparison": "includes", "target": "resource.subject"}}`,
			expectedValue: RuleMap{"user.patients": &TargetComparison{Comparison: ComparisonIncludes, Target: "resource.subject"}},
		},
		{
			name:          "should parse multivalue comparison",
			json:          `{"user.groups": {"comparison": "superset", "value": ["admin", "doctor"]}}`,
			expectedValue: RuleMap{"user.groups": &MultivalueComparison{Comparison: ComparisonSuperset, Values: []string{"admin", "doctor"}}},
		},
		{
			name:          "should parse value comparison",
			json:          `{"user.id": {"comparison": "notEquals", "value": "bob"}}`,
			expectedValue: RuleMap{"user.id": &ValueComparison{Comparison: ComparisonNotEquals, Value: "bob"}},
		},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			var value RuleMap

			if err := json.Unmarshal([]byte(fixture.json), &value); err != nil {
				if fixture.expectedErr == "" {
					t.Errorf("unexpected error: %s", err)
					return
				}
				assert.ErrorContains(t, err, fixture.expectedErr)
			} else if fixture.expectedErr != "" {
				t.Errorf("expected error matching: %s; got nil", fixture.expectedErr)
				return
			}

			assert.Equal(t, fixture.expectedValue, value)
		})
	}
}

func TestPolicyRules_UnmarshalJSON(t *testing.T) {
	for _, fixture := range []struct {
		name          string
		json          string
		expectedValue PolicyRules
		expectedErr   string
	}{
		{
			name: "should handle static bool value",
			// Anyone can read data.
			expectedValue: PolicyRules{"readData": StaticRule(true)},
			json:          `{"readData": true}`,
		},
		{
			name: "should handle many rules",
			expectedValue: PolicyRules{"writeData": RuleMappings{
				{"user.patients": &TargetComparison{Comparison: ComparisonIncludes, Target: "resource.subject"}},
				{"user.groups": &MultivalueComparison{Comparison: ComparisonSuperset, Values: []string{"admin", "doctor"}}},
				{"user.id": &ValueComparison{Comparison: ComparisonNotEquals, Value: "bob"}},
			}, "readData": StaticRule(true)},
			// Users in admin or doctor groups can view their patient's data, except for bob.
			json: `{"writeData": [
				{"user.patients": {"comparison": "includes", "target": "resource.subject"}},
				{"user.groups": {"comparison": "superset", "value": ["admin", "doctor"]}},
				{"user.id": {"comparison": "notEquals", "value": "bob"}}
			], "readData": true}`,
		},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			var value PolicyRules

			if err := json.Unmarshal([]byte(fixture.json), &value); err != nil {
				if fixture.expectedErr == "" {
					t.Errorf("unexpected error: %s", err)
					return
				}
				assert.ErrorContains(t, err, fixture.expectedErr)
			} else if fixture.expectedErr != "" {
				t.Errorf("expected error matching: %s; got nil", fixture.expectedErr)
				return
			}

			assert.Equal(t, fixture.expectedValue, value)
		})
	}
}
