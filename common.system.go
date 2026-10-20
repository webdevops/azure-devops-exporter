package main

import (
	"context"

	"github.com/webdevops/go-common/system"
)

func initSystem() {
	system.AutoProcMemLimit(context.Background(), logger.Logger)
}
