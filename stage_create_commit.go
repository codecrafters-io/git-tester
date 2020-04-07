package main

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"
)

func testCreateCommit(executable *Executable, logger *customLogger) error {
	tempDir, err := ioutil.TempDir("", "worktree")
	if err != nil {
		return err
	}

	executable.WorkingDir = tempDir

	logger.Debugf("Running git init")
	runGitCmd(tempDir, "init")

	logger.Debugf("Creating some files & directories")

	rootFile := "root.txt"
	firstLevel := randomStringsShort(3)
	rootFile, rootDir1, rootDir2 := firstLevel[0], firstLevel[1], firstLevel[2]
	secondLevel := randomStringsShort(2)
	rootDir1File1, rootDir1File2 := secondLevel[0], secondLevel[1]
	thirdLevel := randomStringsShort(2)
	rootDir2File1, rootDir2File2 := thirdLevel[0], thirdLevel[1]

	writeFile(tempDir, rootFile)
	writeFile(tempDir, path.Join(rootDir1, rootDir1File1))
	writeFile(tempDir, path.Join(rootDir1, rootDir1File2))
	writeFile(tempDir, path.Join(rootDir2, rootDir2File1))

	logger.Debugf("Running git commit --all")
	runGitCmd(tempDir, "commit", "--all", "-m", "initial commit")
	parentCommitSha := strings.TrimSpace(runGitCmd(tempDir, "rev-parse", "@"))

	logger.Debugf("Creating another file")
	writeFile(tempDir, path.Join(rootDir2, rootDir2File2))
	runGitCmd(tempDir, "add", ".")
	logger.Debugf("Creating a new tree object")
	treeSha := strings.TrimSpace(runGitCmd(tempDir, "write-tree"))

	commitMessage := randomString()

	logger.Debugf("Running ./your_git.sh commit-tree <tree_sha> -p <commit_sha> -m <message>")
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
	result, err = runGitCmdUnsafe(tempDir, "cat-file", "commit", sha)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	if err = assertStdoutContains(result, fmt.Sprintf("tree %s", treeSha)); err != nil {
		return err
	}
	if err = assertStdoutContains(result, fmt.Sprintf("parent %s", parentCommitSha)); err != nil {
		return err
	}
	if err = assertStdoutContains(result, commitMessage); err != nil {
		return err
	}

	return nil
}
