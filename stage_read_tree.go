package main

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func testReadTree(executable *Executable, logger *customLogger) error {
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

	logger.Debugf("Writing a tree to git storage..")

	rootFile := randomStringShort()
	rootDir1 := randomStringShort()
	rootDir1File1 := randomStringShort()
	rootDir1File2 := randomStringShort()
	rootDir2 := randomStringShort()
	rootDir2File1 := randomStringShort()

	writeFile(tempDir, rootFile)
	writeFile(tempDir, path.Join(rootDir1, rootDir1File1))
	writeFile(tempDir, path.Join(rootDir1, rootDir1File2))
	writeFile(tempDir, path.Join(rootDir2, rootDir2File1))

	runGitCmd(tempDir, "add", ".")
	stdout := runGitCmd(tempDir, "write-tree")
	sha := strings.TrimSpace(stdout)

	logger.Debugf("Running ./your_git.sh ls-tree --name-only")
	result, err := executable.Run("ls-tree", "--name-only", sha)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	expectedStdout := runGitCmd(tempDir, "ls-tree", "--name-only", sha)
	if err = assertStdout(result, expectedStdout); err != nil {
		return err
	}

	return nil
}

func writeFile(rootDir string, filepath string) {
	filepath = path.Join(rootDir, filepath)
	if err := os.MkdirAll(path.Dir(filepath), 0700); err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(filepath, []byte(randomString()), os.ModePerm); err != nil {
		panic(err)
	}
}
