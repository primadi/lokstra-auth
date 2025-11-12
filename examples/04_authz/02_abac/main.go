package main

import (
	"context"
	"fmt"
	"log"

	subject "github.com/primadi/lokstra-auth/03_subject"
	authz "github.com/primadi/lokstra-auth/04_authz"
	"github.com/primadi/lokstra-auth/04_authz/abac"
)

// SimpleAttributeProvider provides attributes from identity metadata
type SimpleAttributeProvider struct{}

func (p *SimpleAttributeProvider) GetSubjectAttributes(ctx context.Context, subjectID string) (map[string]any, error) {
	return nil, nil // Not used in this example
}

func (p *SimpleAttributeProvider) GetResourceAttributes(ctx context.Context, resource *authz.Resource) (map[string]any, error) {
	return nil, nil // Not used in this example
}

func (p *SimpleAttributeProvider) GetEnvironmentAttributes(ctx context.Context) (map[string]any, error) {
	return nil, nil // Not used in this example
}

func main() {
	fmt.Printf("=== ABAC Authorization Example ===\n")

	ctx := context.Background()

	// Create ABAC evaluator
	provider := &SimpleAttributeProvider{}
	evaluator := abac.NewEvaluator(provider, false)

	// Rule 1: Allow access if user department matches document department
	rule1 := &abac.Rule{
		ID:       "rule-1",
		Effect:   "allow",
		Priority: 10,
		Conditions: []abac.Condition{
			{
				Type:     "subject",
				Key:      "department",
				Operator: "eq",
				Value:    "engineering",
			},
			{
				Type:     "resource",
				Key:      "department",
				Operator: "eq",
				Value:    "engineering",
			},
		},
	}
	evaluator.AddRule(rule1)

	// Rule 2: Allow managers to access all documents
	rule2 := &abac.Rule{
		ID:       "rule-2",
		Effect:   "allow",
		Priority: 20,
		Conditions: []abac.Condition{
			{
				Type:     "subject",
				Key:      "role",
				Operator: "eq",
				Value:    "manager",
			},
		},
	}
	evaluator.AddRule(rule2)

	// Rule 3: Deny access to confidential documents outside business hours
	rule3 := &abac.Rule{
		ID:       "rule-3",
		Effect:   "deny",
		Priority: 30,
		Conditions: []abac.Condition{
			{
				Type:     "resource",
				Key:      "classification",
				Operator: "eq",
				Value:    "confidential",
			},
			{
				Type:     "environment",
				Key:      "time_of_day",
				Operator: "gt",
				Value:    18,
			},
		},
	}
	evaluator.AddRule(rule3)

	fmt.Println("ABAC Rules:")
	fmt.Println("Rule 1: Allow if user department == resource department == engineering")
	fmt.Println("Rule 2: Allow if user role == manager")
	fmt.Println("Rule 3: Deny confidential documents after 6 PM")
	fmt.Println()

	// Test Case 1: Engineer accessing engineering document
	engineerIdentity := &subject.IdentityContext{
		Subject: &subject.Subject{
			ID:   "user-1",
			Type: "user",
		},
		Metadata: map[string]any{
			"department": "engineering",
			"role":       "engineer",
		},
	}

	request1 := &authz.AuthorizationRequest{
		Subject: engineerIdentity,
		Action:  authz.Action("read"),
		Resource: &authz.Resource{
			Type: "document",
			ID:   "doc-123",
			Attributes: map[string]any{
				"department":     "engineering",
				"classification": "internal",
			},
		},
		Context: map[string]any{
			"time_of_day": 14, // 2 PM
		},
	}

	decision1, err := evaluator.Evaluate(ctx, request1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Test 1 - Engineer accessing engineering document:\n")
	fmt.Printf("  Subject: %s (department: %s, role: %s)\n",
		engineerIdentity.Subject.ID,
		engineerIdentity.Metadata["department"],
		engineerIdentity.Metadata["role"])
	fmt.Printf("  Resource: %s:%s (department: %s)\n",
		request1.Resource.Type,
		request1.Resource.ID,
		request1.Resource.Attributes["department"])
	fmt.Printf("  Decision: %v\n", decision1.Allowed)
	fmt.Printf("  Reason: %s\n\n", decision1.Reason)

	// Test Case 2: Manager accessing any document
	managerIdentity := &subject.IdentityContext{
		Subject: &subject.Subject{
			ID:   "user-2",
			Type: "user",
		},
		Metadata: map[string]any{
			"department": "sales",
			"role":       "manager",
		},
	}

	request2 := &authz.AuthorizationRequest{
		Subject: managerIdentity,
		Action:  authz.Action("read"),
		Resource: &authz.Resource{
			Type: "document",
			ID:   "doc-456",
			Attributes: map[string]any{
				"department":     "engineering",
				"classification": "internal",
			},
		},
		Context: map[string]any{
			"time_of_day": 10,
		},
	}

	decision2, err := evaluator.Evaluate(ctx, request2)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Test 2 - Manager accessing engineering document:\n")
	fmt.Printf("  Subject: %s (department: %s, role: %s)\n",
		managerIdentity.Subject.ID,
		managerIdentity.Metadata["department"],
		managerIdentity.Metadata["role"])
	fmt.Printf("  Resource: %s:%s (department: %s)\n",
		request2.Resource.Type,
		request2.Resource.ID,
		request2.Resource.Attributes["department"])
	fmt.Printf("  Decision: %v\n", decision2.Allowed)
	fmt.Printf("  Reason: %s\n\n", decision2.Reason)

	// Test Case 3: Accessing confidential document after hours (should be denied)
	request3 := &authz.AuthorizationRequest{
		Subject: managerIdentity,
		Action:  authz.Action("read"),
		Resource: &authz.Resource{
			Type: "document",
			ID:   "doc-789",
			Attributes: map[string]any{
				"department":     "hr",
				"classification": "confidential",
			},
		},
		Context: map[string]any{
			"time_of_day": 20, // 8 PM
		},
	}

	decision3, err := evaluator.Evaluate(ctx, request3)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Test 3 - Manager accessing confidential document after hours:\n")
	fmt.Printf("  Subject: %s (role: %s)\n",
		managerIdentity.Subject.ID,
		managerIdentity.Metadata["role"])
	fmt.Printf("  Resource: %s:%s (classification: %s)\n",
		request3.Resource.Type,
		request3.Resource.ID,
		request3.Resource.Attributes["classification"])
	fmt.Printf("  Time: %d:00\n", request3.Context["time_of_day"])
	fmt.Printf("  Decision: %v\n", decision3.Allowed)
	fmt.Printf("  Reason: %s\n\n", decision3.Reason)

	// Test Case 4: Engineer from different department
	salesEngineer := &subject.IdentityContext{
		Subject: &subject.Subject{
			ID:   "user-3",
			Type: "user",
		},
		Metadata: map[string]any{
			"department": "sales",
			"role":       "engineer",
		},
	}

	request4 := &authz.AuthorizationRequest{
		Subject: salesEngineer,
		Action:  authz.Action("read"),
		Resource: &authz.Resource{
			Type: "document",
			ID:   "doc-999",
			Attributes: map[string]any{
				"department":     "engineering",
				"classification": "internal",
			},
		},
		Context: map[string]any{
			"time_of_day": 14,
		},
	}

	decision4, err := evaluator.Evaluate(ctx, request4)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Test 4 - Sales engineer accessing engineering document:\n")
	fmt.Printf("  Subject: %s (department: %s, role: %s)\n",
		salesEngineer.Subject.ID,
		salesEngineer.Metadata["department"],
		salesEngineer.Metadata["role"])
	fmt.Printf("  Resource: %s:%s (department: %s)\n",
		request4.Resource.Type,
		request4.Resource.ID,
		request4.Resource.Attributes["department"])
	fmt.Printf("  Decision: %v\n", decision4.Allowed)
	fmt.Printf("  Reason: %s\n", decision4.Reason)
}
