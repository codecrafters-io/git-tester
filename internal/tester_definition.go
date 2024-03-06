package internal

import (
	"github.com/codecrafters-io/tester-utils/tester_definition"
)

var testerDefinition = tester_definition.TesterDefinition{
	AntiCheatTestCases:    []tester_definition.TestCase{},
	ExecutableFileName: "your_git.sh",
	TestCases: []tester_definition.TestCase{
		{
			Slug:                    "init",
			TestFunc:                testInit,
		},
		{
			Slug:                    "read_blob",
			TestFunc:                testReadBlob,
		},
		{
			Slug:                    "create_blob",
			TestFunc:                testCreateBlob,
		},
		{
			Slug:                    "read_tree",
			TestFunc:                testReadTree,
		},
		{
			Slug:                    "write_tree",
			TestFunc:                testWriteTree,
		},
		{
			Slug:                    "create_commit",
			TestFunc:                testCreateCommit,
		},
		{
			Slug:                    "clone_repository",
			TestFunc:                testCloneRepository,
		},
	},
}
