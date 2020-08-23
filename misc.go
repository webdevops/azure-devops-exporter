package main

import (
	"strconv"
	"time"
)

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func int64ToString(v int64) string {
	return strconv.FormatInt(v, 10)
}

func arrayStringContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func timeToFloat64(v time.Time) float64 {
	return float64(v.Unix())
}
