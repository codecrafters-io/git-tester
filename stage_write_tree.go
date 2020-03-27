package main

import (
	"io/ioutil"
	"path"
	"sort"
	"strings"
)

func testWriteTree(executable *Executable, logger *customLogger) error {
	tempDir, err := ioutil.TempDir("", "worktree")
	if err != nil {
		return err
	}

	executable.WorkingDir = tempDir

	logger.Debugf("Running ./your_git.sh init")
	_, err = executable.Run("init")
	if err != nil {
		return err
	}

	logger.Debugf("Creating some files & directories")

	rootFile := "root.txt"
	firstLevel := randomStringsShort(3)
	rootFile, rootDir1, rootDir2 := firstLevel[0], firstLevel[1], firstLevel[2]
	secondLevel := randomStringsShort(2)
	rootDir1File1, rootDir1File2 := secondLevel[0], secondLevel[1]
	rootDir2File1 := randomStringShort()

	writeFile(tempDir, rootFile)
	writeFile(tempDir, path.Join(rootDir1, rootDir1File1))
	writeFile(tempDir, path.Join(rootDir1, rootDir1File2))
	writeFile(tempDir, path.Join(rootDir2, rootDir2File1))

	// If we're running against git, update the index
	runGitCmd(tempDir, "add", ".")

	logger.Debugf("Running ./your_git.sh write-tree")
	result, err := executable.Run("write-tree")
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	sha := strings.TrimSpace(string(result.Stdout))
	result, err = runGitCmdUnsafe(tempDir, "ls-tree", "--name-only", sha)
	if err != nil {
		return err
	}
	expectedValues := []string{rootFile, rootDir1, rootDir2}
	sort.Strings(expectedValues)
	expectedStdout := strings.Join(
		expectedValues,
		"\n",
	) + "\n"
	if err = assertStdout(result, expectedStdout); err != nil {
		return err
	}

	return nil
}
