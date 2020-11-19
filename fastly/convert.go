package fastly

func uintOrDefault(int *uint) uint {
	if int == nil {
		return 0
	}
	return *int
}
