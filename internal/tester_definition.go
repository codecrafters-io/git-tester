package internal

import (
	testerutils "github.com/codecrafters-io/tester-utils"
)

var testerDefinition = testerutils.TesterDefinition{
	AntiCheatStages:    []testerutils.Stage{},
	ExecutableFileName: "your_git.sh",
	Stages: []testerutils.Stage{
		{
			Slug:                    "init",
			Title:                   "Initialize the .git directory",
			TestFunc:                testInit,
			ShouldRunPreviousStages: true,
		},
		{
			Slug:                    "read_blob",
			Title:                   "Read a blob object",
			TestFunc:                testReadBlob,
			ShouldRunPreviousStages: true,
		},
		{
			Slug:                    "create_blob",
			Title:                   "Create a blob object",
			TestFunc:                testCreateBlob,
			ShouldRunPreviousStages: true,
		},
		{
			Slug:                    "read_tree",
			Title:                   "Read a tree object",
			TestFunc:                testReadTree,
			ShouldRunPreviousStages: true,
		},
		{
			Slug:                    "write_tree",
			Title:                   "Write a tree object",
			TestFunc:                testWriteTree,
			ShouldRunPreviousStages: true,
		},
		{
			Slug:                    "create_commit",
			Title:                   "Create a commit",
			TestFunc:                testCreateCommit,
			ShouldRunPreviousStages: true,
		},
		{
			Slug:                    "clone_repository",
			Title:                   "Clone a repository",
			TestFunc:                testCloneRepository,
			ShouldRunPreviousStages: true,
		},
	},
}
