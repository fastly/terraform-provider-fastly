package fastly

import "testing"

func TestValidateLoggingFormatVersion(t *testing.T) {
	validVersions := []int{
		1,
		2,
	}
	for _, v := range validVersions {
		_, errors := validateLoggingFormatVersion(v, "format_version")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid format version: %q", v, errors)
		}
	}

	invalidVersions := []int{
		0,
		3,
		4,
		5,
	}
	for _, v := range invalidVersions {
		_, errors := validateLoggingFormatVersion(v, "format_version")
		if len(errors) != 1 {
			t.Fatalf("%q should not be a valid format version", v)
		}
	}
}

func TestValidateLoggingMessageType(t *testing.T) {
	validTypes := []string{
		"classic",
		"loggly",
		"logplex",
		"blank",
	}
	for _, v := range validTypes {
		_, errors := validateLoggingMessageType(v, "message_type")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid message type: %q", v, errors)
		}
	}

	invalidTypes := []string{
		"invalid_type_1",
		"invalid_type_2",
	}
	for _, v := range invalidTypes {
		_, errors := validateLoggingMessageType(v, "message_type")
		if len(errors) != 1 {
			t.Fatalf("%q should not be a valid message type", v)
		}
	}
}

func TestValidateLoggingPlacement(t *testing.T) {
	validPlacements := []string{
		"none",
		"waf_debug",
	}
	for _, v := range validPlacements {
		_, errors := validateLoggingPlacement(v, "placement")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid placement", v)
		}
	}

	invalidPlacements := []string{
		"invalid_placement_1",
		"invalid_placement_2",
	}
	for _, v := range invalidPlacements {
		_, errors := validateLoggingPlacement(v, "placement")
		if len(errors) != 1 {
			t.Fatalf("%q should not be a valid placement", v)
		}
	}
}

func TestValidateDirectorQuorum(t *testing.T) {
	validQuorums := []int{
		0,
		9,
		55,
		83,
		100,
	}
	for _, v := range validQuorums {
		_, errors := validateDirectorQuorum(v, "quorum")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid director quorum: %q", v, errors)
		}
	}

	invalidQuorums := []int{
		-1,
		-50,
		101,
		150,
	}
	for _, v := range invalidQuorums {
		_, errors := validateDirectorQuorum(v, "quorum")
		if len(errors) != 1 {
			t.Fatalf("%q should not be a valid director quorum", v)
		}
	}
}

func TestValidateDirectorType(t *testing.T) {
	validVersions := []int{
		1,
		3,
		4,
	}
	for _, v := range validVersions {
		_, errors := validateDirectorType(v, "type")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid director type: %q", v, errors)
		}
	}

	invalidVersions := []int{
		0,
		-1,
		2,
		5,
		6,
	}
	for _, v := range invalidVersions {
		_, errors := validateDirectorType(v, "type")
		if len(errors) != 1 {
			t.Fatalf("%q should not be a valid director type", v)
		}
	}
}

func TestValidateConditionType(t *testing.T) {
	for _, ctvt := range []struct {
		value          string
		expectedWarns  int
		expectedErrors int
	}{
		{"REQUEST", 0, 0},
		{"RESPONSE", 0, 0},
		{"CACHE", 0, 0},
		{"request", 0, 1},
		{"response", 0, 1},
		{"cache", 0, 1},
	} {
		t.Run(ctvt.value, func(t *testing.T) {
			actualWarns, actualErrors := validateConditionType()(ctvt.value, "type")
			if len(actualWarns) != ctvt.expectedWarns {
				t.Errorf("expected %d warnings, actual %d ", ctvt.expectedWarns, len(actualWarns))
			}
			if len(actualErrors) != ctvt.expectedErrors {
				t.Errorf("expected %d errors, actual %d ", ctvt.expectedErrors, len(actualErrors))
			}
		})
	}
}

func TestValidateHeaderAction(t *testing.T) {
	validActions := []string{
		"set",
		"append",
		"delete",
		"regex",
		"regex_repeat",
	}
	for _, v := range validActions {
		_, errors := validateHeaderAction(v, "action")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid header action: %q", v, errors)
		}
	}

	invalidActions := []string{
		"invalid_action_1",
		"invalid_action_2",
	}
	for _, v := range invalidActions {
		_, errors := validateHeaderAction(v, "action")
		if len(errors) != 1 {
			t.Fatalf("%q should not be a valid header action", v)
		}
	}
}

func TestValidateHeaderType(t *testing.T) {
	validTypes := []string{
		"request",
		"fetch",
		"cache",
		"response",
	}
	for _, v := range validTypes {
		_, errors := validateHeaderType(v, "type")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid header type: %q", v, errors)
		}
	}

	invalidTypes := []string{
		"invalid_type_1",
		"invalid_type_2",
	}
	for _, v := range invalidTypes {
		_, errors := validateHeaderType(v, "type")
		if len(errors) != 1 {
			t.Fatalf("%q should not be a valid header type", v)
		}
	}
}

func TestValidateSnippetType(t *testing.T) {
	validTypes := []string{
		"init",
		"recv",
		"hit",
		"miss",
		"pass",
		"fetch",
		"error",
		"deliver",
		"log",
		"none",
	}
	for _, v := range validTypes {
		_, errors := validateSnippetType(v, "type")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid snippet type: %q", v, errors)
		}
	}

	invalidTypes := []string{
		"invalid_type_1",
		"invalid_type_2",
	}
	for _, v := range invalidTypes {
		_, errors := validateSnippetType(v, "type")
		if len(errors) != 1 {
			t.Fatalf("%q should not be a valid snippet type", v)
		}
	}
}
