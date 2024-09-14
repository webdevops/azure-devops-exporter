package main

import (
	"github.com/KimMachineGun/automemlimit/memlimit"
	humanize "github.com/dustin/go-humanize"
)

func initSystem() {
	// set memory limit
	goMemLimit, err := memlimit.SetGoMemLimitWithOpts(
		memlimit.WithProvider(
			memlimit.ApplyFallback(
				memlimit.FromCgroup,
				memlimit.FromSystem,
			),
		),
		memlimit.WithLogger(slogger),
	)

	if goMemLimit > 0 {
		logger.Infof(`GOMEMLIMIT updated to %v`, humanize.Bytes(uint64(goMemLimit)))
	}

	if err != nil {
		logger.Fatal(err)
	}
}
