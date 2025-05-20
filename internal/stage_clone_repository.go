package internal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"

	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type TestFile struct {
	path     string
	contents string
}

type TestRepo struct {
	url            string
	exampleCommits []string
	exampleFiles   []TestFile
}

func (r TestRepo) randomCommit() string {
	return r.exampleCommits[random.RandomInt(0, len(r.exampleCommits))]
}

func (r TestRepo) randomFile() TestFile {
	return r.exampleFiles[random.RandomInt(0, len(r.exampleFiles))]
}

var testRepos = []TestRepo{
	{
		url: "https://github.com/codecrafters-io/git-sample-1",
		exampleCommits: []string{
			"3b0466d22854e57bf9ad3ccf82008a2d3f199550",
		},
		exampleFiles: []TestFile{
			{
				path:     "scooby/dooby/doo",
				contents: "dooby yikes dumpty scooby monkey donkey horsey humpty vanilla doo",
			},
		},
	},
	{
		url: "https://github.com/codecrafters-io/git-sample-2",
		exampleCommits: []string{
			"b521b9179412d90a893bc36f33f5dcfd987105ef",
		},
		exampleFiles: []TestFile{
			{
				path:     "humpty/vanilla/yikes",
				contents: "scooby yikes dooby",
			},
		},
	},
	{
		url: "https://github.com/codecrafters-io/git-sample-3",
		exampleCommits: []string{
			"23f0bc3b5c7c3108e41c448f01a3db31e7064bbb",
			"b521b9179412d90a893bc36f33f5dcfd987105ef",
		},
		exampleFiles: []TestFile{
			{
				path:     "donkey/donkey/monkey",
				contents: "monkey humpty doo scooby dumpty donkey vanilla horsey dooby",
			},
		},
	},
}

func randomRepo() TestRepo {
	index := random.RandomInt(0, len(testRepos))
	return testRepos[index]
}

func testCloneRepository(harness *test_case_harness.TestCaseHarness) error {
	logger := harness.Logger
	executable := harness.Executable

	tempDir, err := os.MkdirTemp("", "worktree")
	if err != nil {
		return err
	}

	executable.WorkingDir = tempDir

	testRepo := randomRepo()

	logger.Infof("$ ./%s clone %s <testDir>", path.Base(executable.Path), testRepo.url)

	oldGitPath, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git executable not found: %v", err)
	}
	oldGitDir := path.Dir(oldGitPath)
	logger.Debugf("Found git executable at: %s", oldGitPath)

	tmpGitDir, err := os.MkdirTemp("/tmp", "git-*")
	if err != nil {
		return err
	}
	logger.Debugf("Created temporary directory for git clone: %s", tmpGitDir)
	tmpGitPath := path.Join(tmpGitDir, "git")
	defer os.RemoveAll(tmpGitDir)

	// Copy the custom_executable to the output path
	command := fmt.Sprintf("sudo mv %s %s", oldGitPath, tmpGitDir)
	fmt.Println(command)
	copyCmd := exec.Command("sh", "-c", command)
	copyCmd.Stdout = os.Stdout
	copyCmd.Stderr = os.Stderr
	if err := copyCmd.Run(); err != nil {
		return fmt.Errorf("CodeCrafters Internal Error: mv1 failed: %w", err)
	}
	logger.Debugf("mv-ed git to temp directory: %s", tmpGitDir)

	defer func() error {
		// Copy the custom_executable to the output path
		command := fmt.Sprintf("sudo mv %s %s", tmpGitPath, oldGitDir)
		fmt.Println(command)
		copyCmd := exec.Command("sh", "-c", command)
		copyCmd.Stdout = io.Discard
		copyCmd.Stderr = io.Discard
		if err := copyCmd.Run(); err != nil {
			return fmt.Errorf("CodeCrafters Internal Error: mv2 failed: %w", err)
		}
		logger.Debugf("mv-ed git to og directory: %s", oldGitDir)

		return nil
	}()

	defer func() error {
		fmt.Println(tmpGitDir)
		// if err := os.RemoveAll(tmpGitDir); err != nil {
		// 	return fmt.Errorf("CodeCrafters Internal Error: delete directory failed: %s", tmpGitDir)
		// }
		return nil
	}()

	result, err := executable.Run("clone", testRepo.url, "test_dir")
	if err != nil {
		return err
	}

	if err = assertExitCode(result, 0); err != nil {
		return err
	}

	repoDir := path.Join(tempDir, "test_dir")
	r, err := git.PlainOpen(repoDir)
	if err != nil {
		return err
	}

	// Test a commit
	commit_sha := testRepo.randomCommit()

	logger.Infof("$ git cat-file commit %s", commit_sha)

	commit, err := r.CommitObject(plumbing.NewHash(commit_sha))
	if err != nil {
		return err
	}

	expected, actual := "Paul Kuruvilla", commit.Author.Name
	if expected != actual {
		return fmt.Errorf("Expected %q as author name, got: %q", expected, actual)
	}
	logger.Successf("Commit contents verified")

	// Test a file
	testFile := testRepo.randomFile()

	logger.Debugf("Reading contents of a sample file")
	bytes, err := os.ReadFile(path.Join(repoDir, testFile.path))
	if err != nil {
		return err
	}

	expected, actual = testFile.contents, string(bytes)
	if expected != actual {
		return fmt.Errorf("Expected %q as file contents, got: %q", expected, actual)
	}
	logger.Successf("File contents verified")

	return nil
}
