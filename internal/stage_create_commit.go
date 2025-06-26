package internal

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func testCreateCommit(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	executable := harness.Executable

	gitTempDir, err := MoveGitToTemp(logger)
	if err != nil {
		return err
	}
	defer gitTempDir.RestoreGit()

	tempDir, err := os.MkdirTemp("", "worktree")
	if err != nil {
		return err
	}

	executable.WorkingDir = tempDir

	logger.Debugf("Running git init")
	if _, err := git.PlainInit(tempDir, false); err != nil {
		return err
	}

	logger.Debugf("Creating some files & directories")

	rootFile := "root.txt"
	firstLevel := random.RandomWords(3)
	rootFile, rootDir1, rootDir2 := firstLevel[0], firstLevel[1], firstLevel[2]
	secondLevel := random.RandomWords(2)
	rootDir1File1, rootDir1File2 := secondLevel[0], secondLevel[1]
	thirdLevel := random.RandomWords(2)
	rootDir2File1, rootDir2File2 := thirdLevel[0], thirdLevel[1]

	writeFile(tempDir, rootFile)
	writeFile(tempDir, path.Join(rootDir1, rootDir1File1))
	writeFile(tempDir, path.Join(rootDir1, rootDir1File2))
	writeFile(tempDir, path.Join(rootDir2, rootDir2File1))

	logger.Debugf("Running git commit --all")
	repository, err := git.PlainOpen(tempDir)
	if err != nil {
		return err
	}

	w, err := repository.Worktree()
	if err != nil {
		return err
	}

	if err = w.AddGlob("."); err != nil {
		return err
	}

	commitHash, err := w.Commit("test2", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}
	parentCommitSha := commitHash.String()

	logger.Debugf("Creating another file")
	writeFile(tempDir, path.Join(rootDir2, rootDir2File2))

	if err = w.AddGlob("."); err != nil {
		return err
	}

	nextCommitHash, err := w.Commit("test2", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	nextCommit, err := repository.CommitObject(nextCommitHash)
	if err != nil {
		return err
	}

	treeSha := nextCommit.TreeHash.String()

	commitMessage := random.RandomString()
	logger.Infof("$ ./%s commit-tree <tree_sha> -p <commit_sha> -m <message>", path.Base(executable.Path))
	result, err := executable.Run("commit-tree", treeSha, "-p", parentCommitSha, "-m", commitMessage)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	sha := strings.TrimSpace(string(result.Stdout))
	if len(sha) != 40 {
		return fmt.Errorf("Expected a 40-char SHA as output. Got: %v", sha)
	}

	logger.Debugf("Running git cat-file commit <sha>")
	c, err := repository.CommitObject(plumbing.NewHash(sha))
	if err != nil {
		return err
	}

	expected, actual := treeSha, c.TreeHash.String()
	if expected != actual {
		return fmt.Errorf("Expected %q as tree, got: %q", expected, actual)
	}

	parent, err := c.Parent(0)
	if err != nil {
		return err
	}
	expected, actual = parentCommitSha, parent.Hash.String()
	if expected != actual {
		return fmt.Errorf("Expected %q as parent commit, got: %q", expected, actual)
	}

	expected, actual = commitMessage+"\n", c.Message
	if expected != actual {
		return fmt.Errorf("Expected %q as commit message, got: %q", expected, actual)
	}

	return nil
}
