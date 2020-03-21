package main

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func testCreateBlob(executable *Executable, logger *customLogger) error {
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

	logger.Debugf("Writing sample file")
	sampleFile := path.Join(tempDir, "test.txt")
	err = ioutil.WriteFile(
		path.Join(tempDir, "test.txt"),
		[]byte("testing string"),
		os.ModePerm,
	)
	if err != nil {
		return err
	}

	logger.Debugf("Running ./your_git.sh hash-object -v <file>")
	result, err := executable.Run("hash-object", "-v", sampleFile)
	if err != nil {
		return err
	}

	expectedSha := strings.TrimSpace(runGitCmd(tempDir, "hash-object", sampleFile))
	if err = assertStdoutContains(result, expectedSha); err != nil {
		return err
	}

	logger.Debugf("Running git cat-file -p <sha>")
	result, err = runGitCmdUnsafe(tempDir, "cat-file", "-p", expectedSha)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	return assertStdout(result, "testing string")
}

func runGitCmdUnsafe(workingDir string, args ...string) (ExecutableResult, error) {
	executable := NewExecutable("git")
	executable.WorkingDir = workingDir
	return executable.Run(args...)
}

func runGitCmd(workingDir string, args ...string) string {
	executable := NewExecutable("git")
	result, err := executable.Run(args...)
	if err != nil {
		panic(err)
	}
	return string(result.Stdout)
}
