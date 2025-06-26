package internal

import (
	"fmt"
	"os"
	"path"

	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-git/v5/storage/filesystem/dotgit"
)

func testReadBlob(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	executable := harness.Executable

	_, err := MoveGitToTemp(harness, logger)
	if err != nil {
		return err
	}

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

	sampleFile := path.Join(tempDir, fmt.Sprintf("%s.txt", random.RandomWord()))
	sampleFileContents := random.RandomString()
	err = os.WriteFile(
		sampleFile,
		[]byte(sampleFileContents),
		os.ModePerm,
	)
	if err != nil {
		return err
	}
	expectedSha := plumbing.ComputeHash(plumbing.BlobObject, []byte(sampleFileContents))

	storage := filesystem.NewObjectStorage(
		dotgit.New(osfs.New(path.Join(tempDir, ".git"))),
		cache.NewObjectLRU(0),
	)
	obj := storage.NewEncodedObject()
	obj.SetType(plumbing.BlobObject)
	writer, err := obj.Writer()
	if err != nil {
		return err
	}

	if _, err := writer.Write([]byte(sampleFileContents)); err != nil {
		return err
	}

	hash, err := storage.SetEncodedObject(obj)
	if err != nil {
		return err
	}

	if hash != expectedSha {
		panic("Expected sha doesn't match!")
	}

	logger.Infof("Added blob object to .git/objects: %s", expectedSha.String())

	logger.Infof("$ ./%s cat-file -p %s", path.Base(executable.Path), expectedSha.String())
	result, err := executable.Run("cat-file", "-p", expectedSha.String())
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	if err = assertStdout(result, sampleFileContents); err != nil {
		return err
	}

	logger.Successf("Output is valid.")
	return nil
}
