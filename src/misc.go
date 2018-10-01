package main


func boolToString(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func arrayInt64Contains(s []int64, e int64) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
