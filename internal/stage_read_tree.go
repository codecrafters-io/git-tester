package internal

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-git/v5/storage/filesystem/dotgit"
)

func testReadTree(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	executable := harness.Executable
	MoveGitToTemp(harness, logger)

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

	logger.Debugf("Writing a tree to git storage..")

	rootFile := "root.txt"
	firstLevel := random.RandomWords(3)
	rootFile, rootDir1, rootDir2 := firstLevel[0], firstLevel[1], firstLevel[2]
	secondLevel := random.RandomWords(2)
	rootDir1File1, rootDir1File2 := secondLevel[0], secondLevel[1]
	rootDir2File1 := random.RandomWord()

	writeFile(tempDir, rootFile)
	writeFile(tempDir, filepath.Join(rootDir1, rootDir1File1))
	writeFile(tempDir, filepath.Join(rootDir1, rootDir1File2))
	writeFile(tempDir, filepath.Join(rootDir2, rootDir2File1))

	repository, err := git.PlainOpen(tempDir)
	if err != nil {
		return err
	}

	w, err := repository.Worktree()
	if err != nil {
		return err
	}

	if err = w.AddGlob("."); err != nil {
		return err
	}

	commitHash, err := w.Commit("test", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	commit, err := repository.CommitObject(commitHash)
	if err != nil {
		return err
	}

	sha := commit.TreeHash.String()
	logger.Infof("$ ./%s ls-tree --name-only %s", path.Base(executable.Path), sha)
	result, err := executable.Run("ls-tree", "--name-only", sha)
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	storage := filesystem.NewObjectStorage(
		dotgit.New(osfs.New(filepath.Join(tempDir, ".git"))),
		cache.NewObjectLRU(0),
	)
	tree, err := object.GetTree(storage, plumbing.NewHash(sha))
	if err != nil {
		return err
	}

	expected := ""

	for _, entry := range tree.Entries {
		expected += entry.Name
		expected += "\n"
	}

	actual := string(result.Stdout)

	if expected != actual {
		return fmt.Errorf("Expected %q as stdout, got: %q", expected, actual)
	}

	return nil
}

func writeFile(rootDir string, filepath string) {
	writeFileContent(random.RandomString(), rootDir, filepath)
}

func writeFileContent(content string, path ...string) {
	filePath := filepath.Join(path...)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		panic(err)
	}
}
