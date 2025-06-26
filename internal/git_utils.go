package internal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"

	"github.com/codecrafters-io/tester-utils/logger"
)

type GitTempDir struct {
	TempDir     string
	OriginalDir string
	TempGitPath string
	logger      *logger.Logger
}

// MoveGitToTemp moves the system git binary to a temporary directory
func MoveGitToTemp(logger *logger.Logger) (*GitTempDir, error) {
	oldGitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("git executable not found: %v", err)
	}
	oldGitDir := path.Dir(oldGitPath)
	logger.Debugf("Found git executable at: %s", oldGitPath)

	tmpGitDir, err := os.MkdirTemp("/tmp", "git-*")
	if err != nil {
		return nil, err
	}
	logger.Debugf("Created temporary directory for git clone: %s", tmpGitDir)
	tmpGitPath := path.Join(tmpGitDir, "git")

	command := fmt.Sprintf("sudo mv %s %s", oldGitPath, tmpGitDir)
	logger.Debugf("command: %s", command)
	moveCmd := exec.Command("sh", "-c", command)
	moveCmd.Stdout = os.Stdout
	moveCmd.Stderr = os.Stderr
	if err := moveCmd.Run(); err != nil {
		os.RemoveAll(tmpGitDir)
		return nil, fmt.Errorf("CodeCrafters Internal Error: mv git to tmp directory failed: %w", err)
	}
	logger.Debugf("mv-ed git to temp directory: %s", tmpGitDir)

	return &GitTempDir{
		TempDir:     tmpGitDir,
		OriginalDir: oldGitDir,
		TempGitPath: tmpGitPath,
		logger:      logger,
	}, nil
}

// RestoreGit moves the git binary back to its original location and cleans up
func (g *GitTempDir) RestoreGit() error {
	command := fmt.Sprintf("sudo mv %s %s", g.TempGitPath, g.OriginalDir)
	g.logger.Debugf("command: %s", command)
	moveCmd := exec.Command("sh", "-c", command)
	moveCmd.Stdout = io.Discard
	moveCmd.Stderr = io.Discard
	if err := moveCmd.Run(); err != nil {
		return fmt.Errorf("CodeCrafters Internal Error: mv restore for git failed: %w", err)
	}
	g.logger.Debugf("mv-ed git to original directory: %s", g.OriginalDir)

	if err := os.RemoveAll(g.TempDir); err != nil {
		return fmt.Errorf("CodeCrafters Internal Error: delete tmp git directory failed: %s", g.TempDir)
	}

	return nil
}
