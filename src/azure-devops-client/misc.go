package AzureDevopsClient

import (
	"strconv"
)

func int64ToString(v int64) string {
	return strconv.FormatInt(v, 10)
}
