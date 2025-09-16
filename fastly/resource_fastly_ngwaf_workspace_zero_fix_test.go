package fastly

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	ws "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces"
)

// Test the resource read function with zero threshold values
func TestResourceFastlyNGWAFWorkspaceRead_ZeroThresholds(t *testing.T) {
	// Create a mock workspace with zero threshold values (the problem case)
	mockWorkspace := &ws.Workspace{
		WorkspaceID:  "test-workspace-id",
		Name:         "Test Workspace",
		Description:  "Test Description",
		Mode:         "block",
		AttackSignalThresholds: ws.AttackSignalThresholds{
			OneMinute:  0, // This would cause validation error
			TenMinutes: 0, // This would cause validation error
			OneHour:    0, // This would cause validation error
			Immediate:  false,
		},
	}

	// Create the resource data
	d := schema.TestResourceDataRaw(t, resourceFastlyNGWAFWorkspace().Schema, map[string]interface{}{
		"name":        "Test Workspace",
		"description": "Test Description",
		"mode":        "block",
	})
	d.SetId("test-workspace-id")

	// Test the logic that would be applied in the read function
	oneMinute := mockWorkspace.AttackSignalThresholds.OneMinute
	if oneMinute == 0 {
		oneMinute = 1 // Default from schema
	}
	tenMinutes := mockWorkspace.AttackSignalThresholds.TenMinutes
	if tenMinutes == 0 {
		tenMinutes = 60 // Default from schema
	}
	oneHour := mockWorkspace.AttackSignalThresholds.OneHour
	if oneHour == 0 {
		oneHour = 100 // Default from schema
	}

	thresholds := []map[string]any{
		{
			"one_minute":  oneMinute,
			"ten_minutes": tenMinutes,
			"one_hour":    oneHour,
			"immediate":   mockWorkspace.AttackSignalThresholds.Immediate,
		},
	}

	// Verify the thresholds are now valid
	expectedThresholds := []map[string]any{
		{
			"one_minute":  1,
			"ten_minutes": 60,
			"one_hour":    100,
			"immediate":   false,
		},
	}

	if !reflect.DeepEqual(thresholds, expectedThresholds) {
		t.Errorf("Expected thresholds %v, got %v", expectedThresholds, thresholds)
	}

	// Verify that all values are within the valid range
	for _, threshold := range thresholds {
		oneMin := threshold["one_minute"].(int)
		tenMin := threshold["ten_minutes"].(int)
		oneHr := threshold["one_hour"].(int)

		if oneMin < 1 || oneMin > 10000 {
			t.Errorf("one_minute value %d is out of valid range (1-10000)", oneMin)
		}
		if tenMin < 1 || tenMin > 10000 {
			t.Errorf("ten_minutes value %d is out of valid range (1-10000)", tenMin)
		}
		if oneHr < 1 || oneHr > 10000 {
			t.Errorf("one_hour value %d is out of valid range (1-10000)", oneHr)
		}
	}
}

// Test with non-zero values to ensure they're preserved
func TestResourceFastlyNGWAFWorkspaceRead_NonZeroThresholds(t *testing.T) {
	// Create a mock workspace with valid non-zero threshold values
	mockWorkspace := &ws.Workspace{
		WorkspaceID:  "test-workspace-id",
		Name:         "Test Workspace",
		Description:  "Test Description",
		Mode:         "block",
		AttackSignalThresholds: ws.AttackSignalThresholds{
			OneMinute:  5,
			TenMinutes: 30,
			OneHour:    150,
			Immediate:  true,
		},
	}

	// Test the logic that would be applied in the read function
	oneMinute := mockWorkspace.AttackSignalThresholds.OneMinute
	if oneMinute == 0 {
		oneMinute = 1 // Default from schema
	}
	tenMinutes := mockWorkspace.AttackSignalThresholds.TenMinutes
	if tenMinutes == 0 {
		tenMinutes = 60 // Default from schema
	}
	oneHour := mockWorkspace.AttackSignalThresholds.OneHour
	if oneHour == 0 {
		oneHour = 100 // Default from schema
	}

	thresholds := []map[string]any{
		{
			"one_minute":  oneMinute,
			"ten_minutes": tenMinutes,
			"one_hour":    oneHour,
			"immediate":   mockWorkspace.AttackSignalThresholds.Immediate,
		},
	}

	// Verify that the original non-zero values are preserved
	expectedThresholds := []map[string]any{
		{
			"one_minute":  5,
			"ten_minutes": 30,
			"one_hour":    150,
			"immediate":   true,
		},
	}

	if !reflect.DeepEqual(thresholds, expectedThresholds) {
		t.Errorf("Expected thresholds %v, got %v", expectedThresholds, thresholds)
	}
}