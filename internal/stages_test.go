package internal

import (
	"os"
	"testing"

	tester_utils_testing "github.com/codecrafters-io/tester-utils/testing"
)

func TestStages(t *testing.T) {
	os.Setenv("CODECRAFTERS_RANDOM_SEED", "1234567890")

	testCases := map[string]tester_utils_testing.TesterOutputTestCase{
		"init_success": {
			UntilStageSlug:      "gg4",
			CodePath:            "./test_helpers/stages/init",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/init",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"init_main": {
			UntilStageSlug:      "gg4",
			CodePath:            "./test_helpers/stages/init_main",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/init_main",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"init_failure": {
			UntilStageSlug:      "gg4",
			CodePath:            "./test_helpers/stages/init_failure",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/init_failure",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"read_blob_success": {
			UntilStageSlug:      "ic4",
			CodePath:            "./test_helpers/stages/read_blob",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/read_blob",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"create_blob_failure": {
			UntilStageSlug:      "jt4",
			CodePath:            "./test_helpers/stages/create_blob_failure",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/create_blob_failure",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"create_blob_no_file": {
			UntilStageSlug:      "jt4",
			CodePath:            "./test_helpers/stages/create_blob_no_file",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/create_blob_no_file",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"create_blob_no_zlib": {
			UntilStageSlug:      "jt4",
			CodePath:            "./test_helpers/stages/create_blob_no_zlib",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/create_blob_no_zlib",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"create_blob_success": {
			UntilStageSlug:      "jt4",
			CodePath:            "./test_helpers/stages/create_blob",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/create_blob",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"read_tree_success": {
			UntilStageSlug:      "kp1",
			CodePath:            "./test_helpers/stages/read_tree",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/read_tree",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"read_tree_exit_code_failure": {
			UntilStageSlug:      "kp1",
			CodePath:            "./test_helpers/stages/read_tree_exit_code_failure",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/read_tree_exit_code_failure",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"write_tree_failure": {
			UntilStageSlug:      "fe4",
			CodePath:            "./test_helpers/stages/write_tree_fail",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/write_tree_fail",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"write_tree_err_not_exist_failure": {
			UntilStageSlug:      "fe4",
			CodePath:            "./test_helpers/stages/write_tree_err_not_exist_fail",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/write_tree_err_not_exist_fail",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"write_tree_success": {
			UntilStageSlug:      "fe4",
			CodePath:            "./test_helpers/stages/write_tree",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/write_tree",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
	}

	tester_utils_testing.TestTesterOutput(t, testerDefinition, testCases)
}

func normalizeTesterOutput(testerOutput []byte) []byte {
	return testerOutput
}
