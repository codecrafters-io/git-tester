package internal

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/codecrafters-io/tester-utils/bytes_diff_visualizer"
	logger "github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func testCreateBlob(harness *test_case_harness.TestCaseHarness) error {
	initRandom()

	logger := harness.Logger
	executable := harness.Executable

	tempDir, err := ioutil.TempDir("", "worktree")
	if err != nil {
		return err
	}

	executable.WorkingDir = tempDir

	logger.Infof("$ ./your_git.sh init")
	_, err = executable.Run("init")
	if err != nil {
		return err
	}

	sampleFileName := fmt.Sprintf("%s.txt", randomStringShort())
	sampleFileContents := randomString()
	logger.Infof("$ echo %q > %s", sampleFileContents, sampleFileName)
	err = ioutil.WriteFile(
		path.Join(tempDir, sampleFileName),
		[]byte(sampleFileContents),
		os.ModePerm,
	)
	if err != nil {
		return err
	}

	expectedSha := plumbing.ComputeHash(plumbing.BlobObject, []byte(sampleFileContents))

	logger.Infof("$ ./your_git.sh hash-object -w %s", sampleFileName)
	result, err := executable.Run("hash-object", "-w", sampleFileName)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	if len(strings.TrimSpace(string(result.Stdout))) != 40 {
		return fmt.Errorf("Expected a 40-char SHA (%q) as output. Got: %q", expectedSha.String(), strings.TrimSpace(string(result.Stdout)))
	}

	actualShaString := strings.TrimSpace(string(result.Stdout))

	if err = assertStdoutContains(result, expectedSha.String()); err != nil {
		printFriendlyBlobFileDiff(logger, tempDir, actualShaString, expectedSha.String(), sampleFileContents)
		return err
	}

	logger.Successf("Output is valid.")

	logger.Infof("$ git cat-file -p %s", expectedSha.String())
	r, err := git.PlainOpen(tempDir)
	if err != nil {
		return err
	}

	blob, err := r.BlobObject(expectedSha)
	if err != nil {
		return err
	}

	blobReader, err := blob.Reader()
	if err != nil {
		return err
	}

	bytes, err := ioutil.ReadAll(blobReader)
	if err != nil {
		return err
	}

	expected, actual := sampleFileContents, string(bytes)

	if expected != actual {
		return fmt.Errorf("Expected %q as file contents, got: %q", expected, actual)
	}

	logger.Successf("Blob file contents are valid.")

	return nil
}

func printFriendlyBlobFileDiff(logger *logger.Logger, repoDir, actualSha, expectedSha, contents string) {
	actualFileRelativePath := path.Join(".git", "objects", actualSha[:2], actualSha[2:])
	actualFilePath := path.Join(repoDir, actualFileRelativePath)
	expectedDecompressedFileContents := []byte("blob " + fmt.Sprint(len(contents)) + "\x00" + contents)
	logger.Infof("Expected SHA: %s", expectedSha)
	logger.Infof("Returned SHA: %s", actualSha)

	actualFileContents, err := os.ReadFile(actualFilePath)
	if err != nil {
		logger.Infof("Note: Did not find file at %q to render diff. Assuming contents are empty.", actualFileRelativePath)

		var in bytes.Buffer
		b := []byte("")
		w := zlib.NewWriter(&in)
		w.Write(b)
		w.Close()

		actualFileContents = in.Bytes()
	}

	compressedActualFileReader, err := zlib.NewReader(bytes.NewReader(actualFileContents))
	if err != nil {
		logger.Infof("Error decompressing file at %q to render diff. Assuming contents are empty.", actualFileRelativePath)
		compressedActualFileReader, _ = zlib.NewReader(bytes.NewReader([]byte{}))
	}

	decompressedActualGitObjectFileContents, err := io.ReadAll(compressedActualFileReader)
	if err != nil {
		logger.Infof("Note: Error decompressing file at %q. Assuming contents are empty.", actualFileRelativePath)
	}

	lines := bytes_diff_visualizer.VisualizeByteDiff(decompressedActualGitObjectFileContents, expectedDecompressedFileContents)
	logger.Errorf("Git object file doesn't match official Git implementation. Diff after zlib decompression:")
	logger.Errorf("")

	for _, line := range lines {
		logger.Plainln(line)
	}

	logger.Errorf("")
}
