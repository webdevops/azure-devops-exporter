package AzureDevopsClient

import (
	"strconv"
	"time"
)

func int64ToString(v int64) string {
	return strconv.FormatInt(v, 10)
}

func parseTime(v string) (*time.Time) {
	for _, layout := range []string{time.RFC3339Nano,time.RFC3339} {
		t, err := time.Parse(layout, v)
		if err == nil {
			return &t
		}
	}

	return nil
}
