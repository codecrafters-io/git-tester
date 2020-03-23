package main

import (
	"io/ioutil"
	"math/rand"
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

	writeFile(tempDir, "file1")
	writeFile(tempDir, "sample_dir_1/file1")
	writeFile(tempDir, "sample_dir_2/file2")
	writeFile(tempDir, "sample_dir_2/file1")

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

	if err = assertStdout(result, "file1\nsample_dir_1\nsample_dir_2\n"); err != nil {
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

func randomString() string {
	words := []string{
		"humpty",
		"dumpty",
		"horsey",
		"donkey",
		"yikes",
		"monkey",
		"doo",
		"scooby",
		"dooby",
		"vanilla",
	}

	return strings.Join(
		[]string{
			words[rand.Intn(10)],
			words[rand.Intn(10)],
			words[rand.Intn(10)],
			words[rand.Intn(10)],
			words[rand.Intn(10)],
			words[rand.Intn(10)],
		},
		" ",
	)
}
