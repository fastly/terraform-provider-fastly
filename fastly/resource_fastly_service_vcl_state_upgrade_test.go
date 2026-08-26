package fastly

import (
	"context"
	"testing"
)

func TestUpgradeServiceVCLStateV0toV1_BotManagementTrue(t *testing.T) {
	// Test upgrading when bot_management was true
	// In rawState, Sets are represented as []any slices
	rawState := map[string]any{
		"product_enablement": []any{
			map[string]any{
				"bot_management": true,
			},
		},
	}

	upgraded, err := upgradeServiceVCLStateV0toV1(context.Background(), rawState, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	peUpgraded, ok := upgraded["product_enablement"].([]any)
	if !ok {
		t.Fatalf("expected product_enablement to be []any, got %T", upgraded["product_enablement"])
	}

	if len(peUpgraded) == 0 {
		t.Fatal("expected product_enablement to have items")
	}

	pe, ok := peUpgraded[0].(map[string]any)
	if !ok {
		t.Fatalf("expected product_enablement element to be map[string]any, got %T", peUpgraded[0])
	}

	botManagement, ok := pe["bot_management"].([]any)
	if !ok {
		t.Fatalf("expected bot_management to be []any, got %T", pe["bot_management"])
	}

	if len(botManagement) != 1 {
		t.Fatalf("expected bot_management to have 1 item, got %d", len(botManagement))
	}

	bmBlock, ok := botManagement[0].(map[string]any)
	if !ok {
		t.Fatalf("expected bot_management element to be map[string]any, got %T", botManagement[0])
	}

	if enabled, ok := bmBlock["enabled"].(bool); !ok || !enabled {
		t.Errorf("expected enabled to be true, got %v", bmBlock["enabled"])
	}

	if contentguard, ok := bmBlock["contentguard"].(string); !ok || contentguard != "off" {
		t.Errorf("expected contentguard to be 'off', got %v", bmBlock["contentguard"])
	}
}

func TestUpgradeServiceVCLStateV0toV1_BotManagementFalse(t *testing.T) {
	// Test upgrading when bot_management was false
	rawState := map[string]any{
		"product_enablement": []any{
			map[string]any{
				"bot_management": false,
			},
		},
	}

	upgraded, err := upgradeServiceVCLStateV0toV1(context.Background(), rawState, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	peUpgraded, ok := upgraded["product_enablement"].([]any)
	if !ok {
		t.Fatalf("expected product_enablement to be []any, got %T", upgraded["product_enablement"])
	}

	if len(peUpgraded) == 0 {
		t.Fatal("expected product_enablement to have items")
	}

	pe, ok := peUpgraded[0].(map[string]any)
	if !ok {
		t.Fatalf("expected product_enablement element to be map[string]any, got %T", peUpgraded[0])
	}

	botManagement, ok := pe["bot_management"].([]any)
	if !ok {
		t.Fatalf("expected bot_management to be []any, got %T", pe["bot_management"])
	}

	// When bot_management was false, it should be converted to an empty list
	if len(botManagement) != 0 {
		t.Fatalf("expected bot_management to be empty, got %d items", len(botManagement))
	}
}

func TestUpgradeServiceVCLStateV0toV1_AlreadyUpgraded(t *testing.T) {
	// Test that already upgraded state (bot_management as []any) is not modified
	// This simulates a state that's already been upgraded or is in transition
	rawState := map[string]any{
		"product_enablement": []any{
			map[string]any{
				"bot_management": []any{
					map[string]any{
						"enabled":      true,
						"contentguard": "on",
					},
				},
			},
		},
	}

	upgraded, err := upgradeServiceVCLStateV0toV1(context.Background(), rawState, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	peUpgraded, ok := upgraded["product_enablement"].([]any)
	if !ok {
		t.Fatalf("expected product_enablement to be []any, got %T", upgraded["product_enablement"])
	}

	if len(peUpgraded) == 0 {
		t.Fatal("expected product_enablement to have items")
	}

	pe, ok := peUpgraded[0].(map[string]any)
	if !ok {
		t.Fatalf("expected product_enablement element to be map[string]any, got %T", peUpgraded[0])
	}

	botManagement, ok := pe["bot_management"].([]any)
	if !ok {
		t.Fatalf("expected bot_management to be []any, got %T", pe["bot_management"])
	}

	if len(botManagement) != 1 {
		t.Fatalf("expected bot_management to have 1 item, got %d", len(botManagement))
	}

	bmBlock, ok := botManagement[0].(map[string]any)
	if !ok {
		t.Fatalf("expected bot_management element to be map[string]any, got %T", botManagement[0])
	}

	if enabled, ok := bmBlock["enabled"].(bool); !ok || !enabled {
		t.Errorf("expected enabled to be true, got %v", bmBlock["enabled"])
	}

	// Should remain "on" since it was already upgraded
	if contentguard, ok := bmBlock["contentguard"].(string); !ok || contentguard != "on" {
		t.Errorf("expected contentguard to remain 'on', got %v", bmBlock["contentguard"])
	}
}

func TestUpgradeServiceVCLStateV0toV1_NoProductEnablement(t *testing.T) {
	// Test that state without product_enablement is not modified
	rawState := map[string]any{
		"name": "test-service",
	}

	upgraded, err := upgradeServiceVCLStateV0toV1(context.Background(), rawState, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, exists := upgraded["product_enablement"]; exists {
		t.Error("expected product_enablement to not be present in upgraded state")
	}

	if name := upgraded["name"]; name != "test-service" {
		t.Errorf("expected name to remain 'test-service', got %v", name)
	}
}

func TestUpgradeServiceVCLStateV0toV1_NilState(t *testing.T) {
	// Test that nil state returns nil without error
	upgraded, err := upgradeServiceVCLStateV0toV1(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if upgraded != nil {
		t.Error("expected nil state to remain nil")
	}
}
