package fastly

// strToPtr returns a pointer to the passed string.
func strToPtr(s string) *string {
	return &s
}

// intToPtr returns a pointer to the passed int.
func intToPtr(i int) *int {
	return &i
}

// boolToPtr returns a pointer to the passed bool.
func boolToPtr(i bool) *bool {
	return &i
}
