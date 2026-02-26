package version

import (
	"log/slog"
	"os/exec"
	"strings"
)

var (
	// Version is the version of this build.
	Version = "0.0.0"
	// Commit is the commit hash of this build.
	Commit = "unknown"
)

const Git = "git"

var (
	GitGetCommitHash = []string{"rev-parse", "--short", "HEAD"}
	GitGetVersion    = []string{"describe", "--tags", "--always"}
)

func init() {
	if !checkGitInstalled() {
		slog.Warn("Git is not installed on this system. The version and commit hash will be invalid.", "package", "version")
		return
	}
	Version = commandOutputOr(Version)(Git, GitGetVersion)
	Commit = commandOutputOr(Commit)(Git, GitGetCommitHash)
}

func commandOutputOr(or string) func(name string, args []string) string {
	return func(name string, args []string) string {
		out, err := exec.Command(name, args...).CombinedOutput()
		if err != nil {
			return or
		}
		return strings.TrimSpace(string(out))
	}
}

func checkGitInstalled() bool {
	_, err := exec.LookPath(Git)
	return err == nil
}
