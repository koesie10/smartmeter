package version

import (
	"fmt"
	"time"
)

var (
	// Version is the version of this build
	Version string
	// Commit is the commit of this build
	Commit string
	// BuildDate is the build date of this build in RFC3339 format
	BuildDate string
	// BuildTime is the build date of this build as a Go time.Time
	BuildTime time.Time
)

func init() {
	if Version == "" {
		Version = "SNAPSHOT"
	}

	if BuildDate == "" {
		BuildDate = time.Time{}.Format(time.RFC3339)
	}

	var err error
	BuildTime, err = time.ParseInLocation(time.RFC3339, BuildDate, time.UTC)
	if err != nil {
		panic(fmt.Sprintf("failed to parse build time: %v", err))
	}
}
