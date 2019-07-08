package fastly

import "testing"

func TestValidateLoggingFormatVersion(t *testing.T) {
	for name, testcase := range map[string]struct {
		value          int
		expectedWarns  int
		expectedErrors int
	}{
		"0": {0, 0, 1},
		"1": {1, 0, 0},
		"2": {2, 0, 0},
		"3": {3, 0, 1},
		"4": {4, 0, 1},
		"5": {5, 0, 1},
	} {
		t.Run(name, func(t *testing.T) {
			actualWarns, actualErrors := validateLoggingFormatVersion()(testcase.value, "format_version")
			if len(actualWarns) != testcase.expectedWarns {
				t.Errorf("expected %d warnings, actual %d ", testcase.expectedWarns, len(actualWarns))
			}
			if len(actualErrors) != testcase.expectedErrors {
				t.Errorf("expected %d errors, actual %d ", testcase.expectedErrors, len(actualErrors))
			}
		})
	}
}

func TestValidateLoggingMessageType(t *testing.T) {
	for _, testcase := range []struct {
		value          string
		expectedWarns  int
		expectedErrors int
	}{
		{"classic", 0, 0},
		{"loggly", 0, 0},
		{"logplex", 0, 0},
		{"blank", 0, 0},
		{"CLASSIC", 0, 1},
		{"LOGGLY", 0, 1},
		{"LOGPLEX", 0, 1},
		{"BLANK", 0, 1},
	} {
		t.Run(testcase.value, func(t *testing.T) {
			actualWarns, actualErrors := validateLoggingMessageType()(testcase.value, "message_type")
			if len(actualWarns) != testcase.expectedWarns {
				t.Errorf("expected %d warnings, actual %d ", testcase.expectedWarns, len(actualWarns))
			}
			if len(actualErrors) != testcase.expectedErrors {
				t.Errorf("expected %d errors, actual %d ", testcase.expectedErrors, len(actualErrors))
			}
		})
	}
}

func TestValidateLoggingPlacement(t *testing.T) {
	for _, testcase := range []struct {
		value          string
		expectedWarns  int
		expectedErrors int
	}{
		{"none", 0, 0},
		{"waf_debug", 0, 0},
		{"NONE", 0, 1},
		{"WAF_DEBUG", 0, 1},
	} {
		t.Run(testcase.value, func(t *testing.T) {
			actualWarns, actualErrors := validateLoggingPlacement()(testcase.value, "placement")
			if len(actualWarns) != testcase.expectedWarns {
				t.Errorf("expected %d warnings, actual %d ", testcase.expectedWarns, len(actualWarns))
			}
			if len(actualErrors) != testcase.expectedErrors {
				t.Errorf("expected %d errors, actual %d ", testcase.expectedErrors, len(actualErrors))
			}
		})
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
	for name, testcase := range map[string]struct {
		value          int
		expectedWarns  int
		expectedErrors int
	}{
		"0": {0, 0, 1},
		"1": {1, 0, 0},
		"2": {2, 0, 1},
		"3": {3, 0, 0},
		"4": {4, 0, 0},
		"5": {5, 0, 1},
	} {
		t.Run(name, func(t *testing.T) {
			actualWarns, actualErrors := validateDirectorType()(testcase.value, "type")
			if len(actualWarns) != testcase.expectedWarns {
				t.Errorf("expected %d warnings, actual %d ", testcase.expectedWarns, len(actualWarns))
			}
			if len(actualErrors) != testcase.expectedErrors {
				t.Errorf("expected %d errors, actual %d ", testcase.expectedErrors, len(actualErrors))
			}
		})
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

func TestValidateHeaderType(t *testing.T) {
	for _, testcase := range []struct {
		value          string
		expectedWarns  int
		expectedErrors int
	}{
		{"request", 0, 0},
		{"fetch", 0, 0},
		{"cache", 0, 0},
		{"response", 0, 0},
		{"REQUEST", 0, 1},
		{"FETCH", 0, 1},
		{"CACHE", 0, 1},
		{"RESPONSE", 0, 1},
	} {
		t.Run(testcase.value, func(t *testing.T) {
			actualWarns, actualErrors := validateHeaderType()(testcase.value, "type")
			if len(actualWarns) != testcase.expectedWarns {
				t.Errorf("expected %d warnings, actual %d ", testcase.expectedWarns, len(actualWarns))
			}
			if len(actualErrors) != testcase.expectedErrors {
				t.Errorf("expected %d errors, actual %d ", testcase.expectedErrors, len(actualErrors))
			}
		})
	}
}

func TestValidateSnippetType(t *testing.T) {
	for _, testcase := range []struct {
		value          string
		expectedWarns  int
		expectedErrors int
	}{
		{"init", 0, 0},
		{"recv", 0, 0},
		{"hit", 0, 0},
		{"miss", 0, 0},
		{"pass", 0, 0},
		{"fetch", 0, 0},
		{"error", 0, 0},
		{"deliver", 0, 0},
		{"log", 0, 0},
		{"none", 0, 0},
		{"INIT", 0, 1},
		{"RECV", 0, 1},
		{"HIT", 0, 1},
		{"MISS", 0, 1},
		{"PASS", 0, 1},
		{"FETCH", 0, 1},
		{"ERROR", 0, 1},
		{"DELIVER", 0, 1},
		{"LOG", 0, 1},
		{"NONE", 0, 1},
	} {
		t.Run(testcase.value, func(t *testing.T) {
			actualWarns, actualErrors := validateSnippetType()(testcase.value, "type")
			if len(actualWarns) != testcase.expectedWarns {
				t.Errorf("expected %d warnings, actual %d ", testcase.expectedWarns, len(actualWarns))
			}
			if len(actualErrors) != testcase.expectedErrors {
				t.Errorf("expected %d errors, actual %d ", testcase.expectedErrors, len(actualErrors))
			}
		})
	}
}

func TestValidateBackendWeight(t *testing.T) {
	for name, testcase := range map[string]struct {
		value          int
		expectedWarns  int
		expectedErrors int
	}{
		"1":   {1, 0, 0},
		"55":  {55, 0, 0},
		"100": {100, 0, 0},
		"0":   {0, 0, 1},
		"101": {101, 0, 1},
		"150": {150, 0, 1},
	} {
		t.Run(name, func(t *testing.T) {
			actualWarns, actualErrors := validateBackendWeight()(testcase.value, "weight")
			if len(actualWarns) != testcase.expectedWarns {
				t.Errorf("expected %d warnings, actual %d ", testcase.expectedWarns, len(actualWarns))
			}
			if len(actualErrors) != testcase.expectedErrors {
				t.Errorf("expected %d errors, actual %d ", testcase.expectedErrors, len(actualErrors))
			}
		})
	}
}
