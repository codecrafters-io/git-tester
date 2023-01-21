package internal

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path"
	"sort"
	"strings"
	"time"
	"unsafe"

	"github.com/codecrafters-io/git-tester/tester"
	tester_utils "github.com/codecrafters-io/tester-utils"
	"github.com/fatih/color"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-git/v5/storage/filesystem/dotgit"
)

func testWriteTree(stageHarness *tester_utils.StageHarness) (err error) {
	logger := stageHarness.Logger

	tempDir, err := ioutil.TempDir("", "git-tester")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}

	defer func() {
		e := os.RemoveAll(tempDir)
		if err == nil && e != nil {
			err = fmt.Errorf("remove temp dir: %w", e)
		}
	}()

	git, err := tester.NewCommand(envOrPanic("CODECRAFTERS_GIT"))
	if err != nil {
		return fmt.Errorf("init git: %w", err)
	}

	prefix := color.New(color.FgYellow).Sprint("[your_program]") + " "
	userLogger := log.New(os.Stdout, prefix, 0)

	userProg, err := tester.NewCommand(
		extractPath(stageHarness.Executable),
		tester.WithLogger(userLogger),
	)
	if err != nil {
		return fmt.Errorf("init user prog: %w", err)
	}

	t, err := tester.New(
		tempDir,
		git,      // canonical command
		userProg, // testable
	)
	if err != nil {
		return fmt.Errorf("make tester: %w", err)
	}

	t.Logger = (*log.Logger)(unsafe.Pointer(logger))

	logger.Debugf("Running ./your_git.sh init")

	_, err = t.Run("init", tester.CheckExitCode)
	if err != nil {
		return err
	}

	seed := time.Now().UnixNano()

	t.Do(func(cmd *tester.Command) error {
		r := rand.New(rand.NewSource(seed))

		content := randomLongStringsRand(4, r)

		root := randomStringsRand(3, r) // file1, dir1, dir2

		dir1 := randomStringsRand(2, r) // file2, file3
		dir2 := randomStringsRand(1, r) // file4

		writeFileContent(content[0], cmd.WorkDir, root[0])
		writeFileContent(content[1], cmd.WorkDir, root[1], dir1[0])
		writeFileContent(content[2], cmd.WorkDir, root[1], dir1[1])
		writeFileContent(content[3], cmd.WorkDir, root[2], dir2[0])

		return nil
	})

	_, err = t.DoOnlyArgs(0, "add", ".")
	if err != nil {
		return err
	}

	logger.Debugf("Running ./your_git.sh write-tree")

	sha, err := t.Run("write-tree",
		tester.CheckExitCode,
	//	tester.CheckOutput,
	)
	if err != nil {
		return err
	}

	logger.Debugf("Running ./your_git.sh ls-tree --name-only %v", string(sha))

	_, err = t.Run("ls-tree", "--name-only", string(sha),
		tester.CheckExitCode,
		tester.CheckOutput,
	)
	if err != nil {
		return err
	}

	return nil
}

func testWriteTree0(stageHarness *tester_utils.StageHarness) error {
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

	rootFile := "root.txt"
	firstLevel := randomStringsShort(3)
	rootFile, rootDir1, rootDir2 := firstLevel[0], firstLevel[1], firstLevel[2]
	secondLevel := randomStringsShort(2)
	rootDir1File1, rootDir1File2 := secondLevel[0], secondLevel[1]
	rootDir2File1 := randomStringShort()

	writeFile(tempDir, rootFile)
	writeFile(tempDir, path.Join(rootDir1, rootDir1File1))
	writeFile(tempDir, path.Join(rootDir1, rootDir1File2))
	writeFile(tempDir, path.Join(rootDir2, rootDir2File1))

	r, err := git.PlainOpen(tempDir)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return nil
	}

	// If we're running against git, update the index
	if err = w.AddGlob("."); err != nil {
		return err
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

	logger.Debugf("Running git ls-tree --name-only <sha>")

	storage := filesystem.NewObjectStorage(
		dotgit.New(osfs.New(path.Join(tempDir, ".git"))),
		cache.NewObjectLRU(0),
	)

	obj, err := storage.EncodedObject(plumbing.TreeObject, plumbing.NewHash(sha))
	if err != nil {
		return fmt.Errorf("not a valid object name (no such object): %s", sha)
	}

	tree, err := object.DecodeTree(storage, obj)
	if err != nil {
		return fmt.Errorf("malformed tree object")
	}

	actual := ""
	for _, entry := range tree.Entries {
		actual += entry.Name
		actual += "\n"
	}

	expectedValues := []string{rootFile, rootDir1, rootDir2}
	sort.Strings(expectedValues)
	expected := strings.Join(
		expectedValues,
		"\n",
	) + "\n"

	if expected != actual {
		return fmt.Errorf("Expected %q as stdout, got: %q", expected, actual)
	}

	return nil
}

func envOrPanic(key string) string {
	res := os.Getenv(key)
	if res == "" {
		panic(key + " is not set")
	}

	return res
}

func extractPath(e *tester_utils.Executable) string {
	type dummy struct {
		Path string
	}

	d := (*dummy)(unsafe.Pointer(e))

	return d.Path
}
