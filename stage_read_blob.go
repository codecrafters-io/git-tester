package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func testReadBlob(executable *Executable, logger *customLogger) error {
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

	logger.Debugf("Running git hash-object -w <file>")
	stdout := runGitCmd(tempDir, "hash-object", "-w", sampleFile)
	sha := strings.TrimSpace(stdout)

	logger.Debugf("Running ./your_git.sh cat-file -p <sha>")
	result, err := executable.Run("cat-file", "-p", sha)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	return assertStdout(result, sampleFileContents)
}
