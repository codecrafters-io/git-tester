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

type GitTempDir struct {
	TempDir     string
	OriginalDir string
	TempGitPath string
	logger      *logger.Logger
}

// MoveGitToTemp moves the system git binary to a temporary directory
func MoveGitToTemp(harness *test_case_harness.TestCaseHarness, logger *logger.Logger) (*GitTempDir, error) {
	oldGitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("CodeCrafters Internal Error: git executable not found: %v", err)
	}
	oldGitDir := path.Dir(oldGitPath)

	tmpGitDir, err := os.MkdirTemp("/tmp", "git-*")
	if err != nil {
		return nil, err
	}
	tmpGitPath := path.Join(tmpGitDir, "git")

	command := fmt.Sprintf("sudo mv %s %s", oldGitPath, tmpGitDir)
	moveCmd := exec.Command("sh", "-c", command)
	moveCmd.Stdout = os.Stdout
	moveCmd.Stderr = os.Stderr
	if err := moveCmd.Run(); err != nil {
		os.RemoveAll(tmpGitDir)
		return nil, fmt.Errorf("CodeCrafters Internal Error: mv git to tmp directory failed: %w", err)
	}

	gitTempDir := &GitTempDir{
		TempDir:     tmpGitDir,
		OriginalDir: oldGitDir,
		TempGitPath: tmpGitPath,
		logger:      logger,
	}

	// Register teardown function to automatically restore git
	harness.RegisterTeardownFunc(func() { gitTempDir.restoreGitInternal() })

	return gitTempDir, nil
}

// RestoreGit moves the git binary back to its original location and cleans up
func (g *GitTempDir) restoreGitInternal() error {
	command := fmt.Sprintf("sudo mv %s %s", g.TempGitPath, g.OriginalDir)
	moveCmd := exec.Command("sh", "-c", command)
	moveCmd.Stdout = io.Discard
	moveCmd.Stderr = io.Discard
	if err := moveCmd.Run(); err != nil {
		return fmt.Errorf("CodeCrafters Internal Error: mv restore for git failed: %w", err)
	}

	if err := os.RemoveAll(g.TempDir); err != nil {
		return fmt.Errorf("CodeCrafters Internal Error: delete tmp git directory failed: %s", g.TempDir)
	}

	return nil
}
