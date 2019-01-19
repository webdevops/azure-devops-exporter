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


func scrapeIntervalStatus(v *time.Duration) (ret string) {
	if v != nil && v.Seconds() > 0 {
		ret = v.String()
	} else {
		ret = "disabled"
	}

	return
}
