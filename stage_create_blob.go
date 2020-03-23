package main

import (
	"fmt"
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
	sampleFile := path.Join(tempDir, fmt.Sprintf("%s.txt", randomStringShort()))
	sampleFileContents := randomString()
	err = ioutil.WriteFile(
		sampleFile,
		[]byte(sampleFileContents),
		os.ModePerm,
	)
	if err != nil {
		return err
	}

	logger.Debugf("Running ./your_git.sh hash-object -w <file>")
	result, err := executable.Run("hash-object", "-w", sampleFile)
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

	return assertStdout(result, sampleFileContents)
}
