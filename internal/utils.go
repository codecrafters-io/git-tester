package internal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"

	"github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

// MoveGitToTemp moves the system git binary to a temporary directory
func MoveGitToTemp(harness *test_case_harness.TestCaseHarness, logger *logger.Logger) {
	oldGitPath, err := exec.LookPath("git")
	if err != nil {
		panic(fmt.Sprintf("CodeCrafters Internal Error: git executable not found: %v", err))
	}
	oldGitDir := path.Dir(oldGitPath)

	tmpGitDir, err := os.MkdirTemp("/tmp", "git-*")
	if err != nil {
		panic(fmt.Sprintf("CodeCrafters Internal Error: create tmp git directory failed: %v", err))
	}
	tmpGitPath := path.Join(tmpGitDir, "git")

	command := fmt.Sprintf("mv %s %s", oldGitPath, tmpGitPath)
	moveCmd := exec.Command("sh", "-c", command)
	moveCmd.Stdout = io.Discard
	moveCmd.Stderr = io.Discard
	if err := moveCmd.Run(); err != nil {
		os.RemoveAll(tmpGitDir)
		panic(fmt.Sprintf("CodeCrafters Internal Error: mv git to tmp directory failed: %v", err))
	}

	// Register teardown function to automatically restore git
	harness.RegisterTeardownFunc(func() { restoreGit(tmpGitPath, oldGitDir) })
}

// RestoreGit moves the git binary back to its original location and cleans up
func restoreGit(newPath string, originalPath string) error {
	command := fmt.Sprintf("mv %s %s", newPath, originalPath)
	moveCmd := exec.Command("sh", "-c", command)
	moveCmd.Stdout = io.Discard
	moveCmd.Stderr = io.Discard
	if err := moveCmd.Run(); err != nil {
		panic(fmt.Sprintf("CodeCrafters Internal Error: mv restore for git failed: %v", err))
	}

	if err := os.RemoveAll(path.Dir(newPath)); err != nil {
		panic(fmt.Sprintf("CodeCrafters Internal Error: delete tmp git directory failed: %s", path.Dir(newPath)))
	}

	return nil
}
