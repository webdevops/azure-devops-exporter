package main

import (
	"github.com/webdevops/go-common/system"
)

func initSystem() {
	system.AutoProcMemLimit(logger.Logger)
}
