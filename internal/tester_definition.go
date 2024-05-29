package internal

import (
	"github.com/codecrafters-io/tester-utils/tester_definition"
)

var testerDefinition = tester_definition.TesterDefinition{
	AntiCheatTestCases: []tester_definition.TestCase{},
	ExecutableFileName: "your_git.sh",
	TestCases: []tester_definition.TestCase{
		{
			Slug:     "gg4",
			TestFunc: testInit,
		},
		{
			Slug:     "ic4",
			TestFunc: testReadBlob,
		},
		{
			Slug:     "jt4",
			TestFunc: testCreateBlob,
		},
		{
			Slug:     "kp1",
			TestFunc: testReadTree,
		},
		{
			Slug:     "fe4",
			TestFunc: testWriteTree,
		},
		{
			Slug:     "jm9",
			TestFunc: testCreateCommit,
		},
		{
			Slug:     "mg6",
			TestFunc: testCloneRepository,
		},
	},
}
