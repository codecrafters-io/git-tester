package internal

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"
	"unsafe"

	"github.com/codecrafters-io/git-tester/tester"
	tester_utils "github.com/codecrafters-io/tester-utils"
	"github.com/fatih/color"
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

	gitpath := envOrPanic("CODECRAFTERS_GIT")

	git, err := tester.NewCommand(gitpath)
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

	t.Logf("Running command: init")

	err = t.Run("init", tester.CheckExitCode)
	if err != nil {
		return err
	}

	seed := time.Now().UnixNano()

	t.Logf("Crafting some files")

	err = t.Do(func(cmd *tester.Command, _ int) error {
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
	if err != nil {
		return fmt.Errorf("craft some files: %w", err)
	}

	err = t.RunI(0, "add", ".")
	if err != nil {
		return err
	}

	t.Logf("Running command: write-tree")

	err = t.Run("write-tree", tester.CheckExitCode)
	if err != nil {
		return err
	}

	sha := t.AllStdouts()

	t.Logf("Running command: ls-tree --name-only %s", sha[1])

	err = t.Do(func(cmd *tester.Command, i int) error {
		return cmd.Run("ls-tree", "--name-only", sha[i])
	}, tester.CheckExitCode, tester.CheckOutput)
	if err != nil {
		return err
	}

	t.Logf("Running crosscheck: ls-tree --name-only %s. We run canonical git on files created by git of yours", sha[1])

	err = t.CrossDo(func(cmd *tester.Command, i int) error {
		return cmd.Run("ls-tree", "--name-only", sha[i])
	}, tester.CheckExitCode, tester.CheckOutput)
	if err != nil {
		return err
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
