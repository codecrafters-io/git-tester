package internal

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/codecrafters-io/git-tester/internal/blob_object_verifier"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testCreateBlob(harness *test_case_harness.TestCaseHarness) error {
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

	sampleFileName := fmt.Sprintf("%s.txt", random.RandomWord())
	sampleFileContents := random.RandomString()
	logger.Infof("$ echo %q > %s", sampleFileContents, sampleFileName)
	err = os.WriteFile(
		path.Join(tempDir, sampleFileName),
		[]byte(sampleFileContents),
		os.ModePerm,
	)
	if err != nil {
		return err
	}

	logger.Infof("$ ./%s hash-object -w %s", path.Base(executable.Path), sampleFileName)
	result, err := executable.Run("hash-object", "-w", sampleFileName)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	blobObjectVerifier := blob_object_verifier.BlobObjectVerifier{
		RawContents: []byte(sampleFileContents),
	}

	expectedSha := blobObjectVerifier.ExpectedSha()

	if len(strings.TrimSpace(string(result.Stdout))) != 40 {
		return fmt.Errorf("Expected a 40-char SHA (%q) as output. Got: %q", blobObjectVerifier.ExpectedSha(), strings.TrimSpace(string(result.Stdout)))
	}

	actualSha := strings.TrimSpace(string(result.Stdout))

	logger.Successf("Output is a 40-char SHA.")

	if err = blobObjectVerifier.VerifyFileContents(logger, tempDir, actualSha); err != nil {
		return err
	}

	logger.Successf("Blob file contents are valid.")

	if actualSha != expectedSha {
		logger.Infof("Hint: Your blob file was valid, but the SHA is incorrect.")
		logger.Infof("      This most likely means that you're passing wrong data to your SHA hash function.")
		logger.Infof("      Make sure you're computing the SHA before zlib-compressing the file contents, not after.")
		logger.Infof("      Make sure you're including the file type and size in the SHA computation.")
		logger.Infof("      In this case, the contents passed to the SHA computation should be: %q", blobObjectVerifier.ExpectedDecompressedFileContents())
		return fmt.Errorf("Expected SHA: %q, got: %q", expectedSha, actualSha)
	}

	logger.Successf("Returned SHA matches expected SHA.")

	return nil
}
