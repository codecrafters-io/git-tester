package internal

import (
	testerutils "github.com/codecrafters-io/tester-utils"
)

var testerDefinition = testerutils.TesterDefinition{
	AntiCheatStages:    []testerutils.Stage{},
	ExecutableFileName: "your_git.sh",
	Stages: []testerutils.Stage{
		{
			Number:                  1,
			Slug:                    "init",
			Title:                   "Initialize the .git directory",
			TestFunc:                testInit,
			ShouldRunPreviousStages: true,
		},
		{
			Number:                  2,
			Slug:                    "read_blob",
			Title:                   "Read a blob object",
			TestFunc:                testReadBlob,
			ShouldRunPreviousStages: true,
		},
		{
			Number:                  3,
			Slug:                    "create_blob",
			Title:                   "Create a blob object",
			TestFunc:                testCreateBlob,
			ShouldRunPreviousStages: true,
		},
		{
			Number:                  4,
			Slug:                    "read_tree",
			Title:                   "Read a tree object",
			TestFunc:                testReadTree,
			ShouldRunPreviousStages: true,
		},
		{
			Number:                  5,
			Slug:                    "write_tree",
			Title:                   "Write a tree object",
			TestFunc:                testWriteTree,
			ShouldRunPreviousStages: true,
		},
		{
			Number:                  6,
			Slug:                    "create_commit",
			Title:                   "Create a commit",
			TestFunc:                testCreateCommit,
			ShouldRunPreviousStages: true,
		},
		{
			Number:                  7,
			Slug:                    "clone_repository",
			Title:                   "Clone a repository",
			TestFunc:                testCloneRepository,
			ShouldRunPreviousStages: true,
		},
	},
}
