package internal

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/codecrafters-io/tester-utils/bytes_diff_visualizer"
	"github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testWriteTree(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	executable := harness.Executable
	RelocateSystemGit(harness, logger)

	tempDir, err := os.MkdirTemp("", "worktree")
	if err != nil {
		return err
	}

	executable.WorkingDir = tempDir

	logger.Infof("$ ./%s init", path.Base(executable.Path))
	_, err = executable.Run("init")
	if err != nil {
		return err
	}

	logger.Infof("Creating some files & directories")

	err = generateFiles(tempDir)
	if err != nil {
		panic(err)
	}

	logger.Infof("$ ./%s write-tree", path.Base(executable.Path))
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
	relativePath, _ := filepath.Rel(executable.WorkingDir, gitObjectFilePath)
	logger.Debugf("Reading file at %v", relativePath)

	gitObjectFileContents, err := os.ReadFile(gitObjectFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("Did you write the tree object? Did not find a file in .git/objects/<first 2 chars of sha>/<remaining chars of sha>")
	} else if err != nil {
		return fmt.Errorf("CodeCrafters internal error. Error reading %v: %v", relativePath, err)
	}

	logger.Successf("Found git object file written at .git/objects/%v/%v.", sha[:2], sha[2:])

	err = checkWithGit(tempDir, logger, sha, gitObjectFileContents)
	if err != nil {
		return err
	}

	logger.Infof("$ git ls-tree --name-only %v", sha)
	_, err = runGit(executable.WorkingDir, "ls-tree", "--name-only", sha)
	if err != nil {
		return err
	}

	return nil
}

// Git commands are run in the same tempDir, overwriting user operations
// We need to create 2 set of dirs, with the exact same contents
func checkWithGit(tempDir string, logger *logger.Logger, actualHash string, actualGitObjectFileContents []byte) error {
	_, err := runGit(tempDir, "init")
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

	// check file contents
	expectedGitObjectFilePath := path.Join(tempDir, ".git", "objects", expectedHash[:2], expectedHash[2:])
	expectedGitObjectFileContents, err := os.ReadFile(expectedGitObjectFilePath)
	if err != nil {
		return fmt.Errorf("CodeCrafters internal error. Error reading %v: %v", expectedGitObjectFilePath, err)
	}

	decompressedActualGitObjectFileContents, err := decodeZlib(actualGitObjectFileContents)
	if err != nil {
		return fmt.Errorf("Git object file doesn't match official Git implementation. This file must be zlib-compressed")
	}

	decompressedExpectedGitObjectFileContents, err := decodeZlib(expectedGitObjectFileContents)
	if err != nil {
		return fmt.Errorf("CodeCrafters internal error. Error decoding zlib-compressed file %v: %v", expectedGitObjectFilePath, err)
	}

	if !bytes.Equal(decompressedExpectedGitObjectFileContents, decompressedActualGitObjectFileContents) {
		lines := bytes_diff_visualizer.VisualizeByteDiff(decompressedActualGitObjectFileContents, decompressedExpectedGitObjectFileContents)
		logger.Errorf("Git object file doesn't match official Git implementation. Diff after zlib decompression:")
		logger.Errorf("")
		for _, line := range lines {
			logger.Plainln(line)
		}
		logger.Errorf("")
		return fmt.Errorf("Git object file doesn't match official Git implementation")
	}

	if expected := expectedHash; expected != actualHash {
		return fmt.Errorf("Expected %q as tree hash, got: %q", expected, actualHash)
	}

	return nil
}

func decodeZlib(data []byte) ([]byte, error) {
	// Create a new zlib reader
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	// Read all the decompressed data
	decompressedData, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return decompressedData, nil
}

func generateFiles(root string) error {
	content := random.RandomStrings(4)

	first := random.RandomWords(3) // file1, dir1, dir2

	dir1 := random.RandomWords(2) // file2, file3
	dir2 := random.RandomWords(1) // file4

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
