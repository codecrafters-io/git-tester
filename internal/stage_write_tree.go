package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	tester_utils "github.com/codecrafters-io/tester-utils"
)

func testWriteTree(stageHarness *tester_utils.StageHarness) error {
	initRandom()

	logger := stageHarness.Logger
	executable := stageHarness.Executable

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

	var seed int64

	if seedStr := os.Getenv("CODECRAFTERS_RANDOM_SEED"); seedStr != "" {
		seedInt, err := strconv.Atoi(seedStr)

		if err != nil {
			panic(err)
		}

		seed = int64(seedInt)
	} else {
		seed = time.Now().UnixNano()
	}

	err = generateFiles(tempDir, seed)
	if err != nil {
		panic(err)
	}

	logger.Debugf("Running ./your_git.sh write-tree")
	result, err := executable.Run("write-tree")
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

	gitObjectFilePath := path.Join(executable.WorkingDir, ".git", "objects", sha[:2], sha[2:])
	logger.Debugf("Reading file at %v", gitObjectFilePath)

	gitObjectFileContents, err := os.ReadFile(gitObjectFilePath)
	if err == os.ErrNotExist {
		return fmt.Errorf("Did you write the tree object? Did not find a file in .git/objects/<first 2 chars of sha>/<remaining chars of sha>")
	} else if err != nil {
		return fmt.Errorf("CodeCrafters internal error. Error reading %v: %v", gitObjectFilePath, err)
	}

	logger.Successf("Found git object file written at .git/objects/%v/%v.", sha[:2], sha[2:])

	logger.Debugf("Running git ls-tree --name-only <sha>")

	tree, err := runGit(executable.WorkingDir, "ls-tree", "--name-only", sha)
	if err != nil {
		return err
	}

	err = checkWithGit(sha, tree, gitObjectFileContents, seed)
	if err != nil {
		return err
	}

	return nil
}

func checkWithGit(actualHash string, tree []byte, actualGitObjectFileContents []byte, seed int64) error {
	tempDir, err := ioutil.TempDir("", "worktree")
	if err != nil {
		return err
	}

	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	err = generateFiles(tempDir, seed)
	if err != nil {
		return err
	}

	_, err = runGit(tempDir, "init")
	if err != nil {
		return err
	}

	_, err = runGit(tempDir, "add", ".")
	if err != nil {
		return err
	}

	expectedHashBytes, err := runGit(tempDir, "write-tree")
	if err != nil {
		return err
	}

	expectedHash := string(bytes.TrimSpace(expectedHashBytes))

	expectedTree, err := runGit(tempDir, "ls-tree", "--name-only", expectedHash)
	if err != nil {
		return err
	}

	// check file list first as it's the friendliest diff to read
	if expected := string(expectedTree); expected != string(tree) {
		return fmt.Errorf("Expected %q as stdout, got: %q", expected, tree)
	}

	// check file contents
	expectedGitObjectFilePath := path.Join(tempDir, ".git", "objects", expectedHash[:2], expectedHash[2:])
	expectedGitObjectFileContents, err := os.ReadFile(expectedGitObjectFilePath)
	if err != nil {
		return fmt.Errorf("CodeCrafters internal error. Error reading %v: %v", expectedGitObjectFilePath, err)
	}

	if !bytes.Equal(expectedGitObjectFileContents, actualGitObjectFileContents) {
		return fmt.Errorf("Expected %q as tree object file contents, got: %q", expectedGitObjectFileContents, actualGitObjectFileContents)
	}

	if expected := expectedHash; expected != actualHash {
		return fmt.Errorf("Expected %q as tree hash, got: %q", expected, actualHash)
	}

	return nil
}

func generateFiles(root string, seed int64) error {
	r := rand.New(rand.NewSource(seed))

	content := randomLongStringsRand(4, r)

	first := randomStringsRand(3, r) // file1, dir1, dir2

	dir1 := randomStringsRand(2, r) // file2, file3
	dir2 := randomStringsRand(1, r) // file4

	writeFileContent(content[0], root, first[0])
	writeFileContent(content[1], root, first[1], dir1[0])
	writeFileContent(content[2], root, first[1], dir1[1])
	writeFileContent(content[3], root, first[2], dir2[0])

	return nil
}

func runGit(wd string, args ...string) ([]byte, error) {
	path := findGit()

	return runCmd(wd, path, args...)
}

func runCmd(wd, path string, args ...string) ([]byte, error) {
	cmd := exec.Command(path, args...)

	cmd.Dir = wd

	var outb, errb bytes.Buffer

	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()

	//	fmt.Printf("run: %v %v\nerr: %v\nout: %s\nstderr: %s\n", path, args, err, outb.Bytes(), errb.Bytes())

	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return nil, fmt.Errorf("%s", errb.Bytes())
	}

	if err != nil {
		panic(err)
	}

	return outb.Bytes(), err
}

func findGit() string {
	fromEnv := os.Getenv("CODECRAFTERS_GIT")

	return choosePath(fromEnv, "/usr/bin/codecrafters-secret-git", "git")
}

func choosePath(paths ...string) string {
	for _, p := range paths {
		path, err := exec.LookPath(p)
		if err == nil {
			return path
		}
	}

	panic("no git found")
}
