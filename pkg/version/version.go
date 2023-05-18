package version

import (
	"fmt"
	"runtime"
)

var (
	gitHash   = "unknown"
	gitBranch = "unknown"
	buildDate = "unknown"
)

func Version() string {
	return fmt.Sprintf("Go Version: %s\nGit Branch: %s\nGitHash: %s\nBuild Date: %s", runtime.Version(), gitBranch, gitHash, buildDate)
}
