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
	for name, testcase := range map[string]struct {
		value          int
		expectedWarns  int
		expectedErrors int
	}{
		"0":   {0, 0, 0},
		"55":  {55, 0, 0},
		"100": {100, 0, 0},
		"-1":  {-1, 0, 1},
		"101": {101, 0, 1},
		"150": {150, 0, 1},
	} {
		t.Run(name, func(t *testing.T) {
			actualWarns, actualErrors := validateDirectorQuorum()(testcase.value, "quorum")
			if len(actualWarns) != testcase.expectedWarns {
				t.Errorf("expected %d warnings, actual %d ", testcase.expectedWarns, len(actualWarns))
			}
			if len(actualErrors) != testcase.expectedErrors {
				t.Errorf("expected %d errors, actual %d ", testcase.expectedErrors, len(actualErrors))
			}
		})
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
	for _, testcase := range []struct {
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
		t.Run(testcase.value, func(t *testing.T) {
			actualWarns, actualErrors := validateConditionType()(testcase.value, "type")
			if len(actualWarns) != testcase.expectedWarns {
				t.Errorf("expected %d warnings, actual %d ", testcase.expectedWarns, len(actualWarns))
			}
			if len(actualErrors) != testcase.expectedErrors {
				t.Errorf("expected %d errors, actual %d ", testcase.expectedErrors, len(actualErrors))
			}
		})
	}
}

func TestValidateHeaderAction(t *testing.T) {
	for _, testcase := range []struct {
		value          string
		expectedWarns  int
		expectedErrors int
	}{
		{"set", 0, 0},
		{"append", 0, 0},
		{"delete", 0, 0},
		{"regex", 0, 0},
		{"regex_repeat", 0, 0},
		{"SET", 0, 1},
		{"APPEND", 0, 1},
		{"DELETE", 0, 1},
		{"REGEX", 0, 1},
		{"REGEX_REPEAT", 0, 1},
	} {
		t.Run(testcase.value, func(t *testing.T) {
			actualWarns, actualErrors := validateHeaderAction()(testcase.value, "action")
			if len(actualWarns) != testcase.expectedWarns {
				t.Errorf("expected %d warnings, actual %d ", testcase.expectedWarns, len(actualWarns))
			}
			if len(actualErrors) != testcase.expectedErrors {
				t.Errorf("expected %d errors, actual %d ", testcase.expectedErrors, len(actualErrors))
			}
		})
	}
}
