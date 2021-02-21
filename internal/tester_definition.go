package internal

import (
	testerutils "github.com/codecrafters-io/tester-utils"
)

var testerDefinition = testerutils.TesterDefinition{
	AntiCheatStages:    []testerutils.Stage{},
	ExecutableFileName: "your_git.sh",
	Stages: []testerutils.Stage{
		{
			Slug:     "init",
			Title:    "Stage 1: Initialize the .git directory",
			TestFunc: testInit,
		},
		{
			Slug:     "read_blob",
			Title:    "Stage 2: Read a blob object",
			TestFunc: testReadBlob,
		},
		{
			Slug:     "create_blob",
			Title:    "Stage 3: Create a blob object",
			TestFunc: testCreateBlob,
		},
		{
			Slug:     "read_tree",
			Title:    "Stage 4: Read a tree object",
			TestFunc: testReadTree,
		},
		{
			Slug:     "write_tree",
			Title:    "Stage 5: Write a tree object",
			TestFunc: testWriteTree,
		},
		{
			Slug:     "create_commit",
			Title:    "Stage 6: Create a commit object",
			TestFunc: testCreateCommit,
		},
		{
			Slug:     "clone_repository",
			Title:    "Stage 7: Clone a repository",
			TestFunc: testCloneRepository,
		},
	},
}
