package version

import (
	"strconv"
	"time"
)

var (
	// Version is the version of this build.
	Version = "0.0.0-undefined"
	// Commit is the commit hash of this build.
	Commit = "undefined"
	// BuildTime is the build time string of this build.
	BuildTime = "0"
)

func GetBuildTime() time.Time {
	i, err := strconv.ParseInt(BuildTime, 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.Unix(i, 0).Local()
}
