package main

import (
	"fmt"

	"math/rand"
	"time"
)

// StageRunnerResult is returned from StageRunner.Run()
type StageRunnerResult struct {
	lastStageIndex int
	error          error
	logger         *customLogger
}

// IsSuccess says whether a StageRunnerResult was successful
// or not
func (res StageRunnerResult) IsSuccess() bool {
	return res.error == nil
}

// StageRunner is used to run multiple stages
type StageRunner struct {
	stages  []Stage
	isDebug bool
}

func newStageRunner(isDebug bool) StageRunner {
	return StageRunner{
		isDebug: isDebug,
		stages: []Stage{
			Stage{
				slug:    "init",
				name:    "Stage 1: Initialize the .git directory",
				logger:  getLogger(isDebug, "[stage-1] "),
				runFunc: testInit,
			},
			Stage{
				slug:    "create_blob",
				name:    "Stage 2: Create a blob object",
				logger:  getLogger(isDebug, "[stage-2] "),
				runFunc: testCreateBlob,
			},
			Stage{
				slug:    "read_blob",
				name:    "Stage 3: Read a blob object",
				logger:  getLogger(isDebug, "[stage-3] "),
				runFunc: testReadBlob,
			},
			Stage{
				slug:    "read_tree",
				name:    "Stage 4: Read a tree object",
				logger:  getLogger(isDebug, "[stage-4] "),
				runFunc: testReadTree,
			},
			Stage{
				slug:    "write_tree",
				name:    "Stage 5: Write a tree object",
				logger:  getLogger(isDebug, "[stage-5] "),
				runFunc: testWriteTree,
			},
		},
	}
}

// Run tests in a specific StageRunner
func (r StageRunner) Run(executable *Executable) StageRunnerResult {
	for index, stage := range r.stages {
		logger := stage.logger
		logger.Infof("Running test: %s", stage.name)

		stageResultChannel := make(chan error, 1)
		go func() {
			err := stage.runFunc(executable, logger)
			stageResultChannel <- err
		}()

		var err error
		select {
		case stageErr := <-stageResultChannel:
			err = stageErr
		case <-time.After(10 * time.Second):
			err = fmt.Errorf("timed out, test exceeded 10 seconds")
		}

		if err != nil {
			reportTestError(err, r.isDebug, logger)
			return StageRunnerResult{
				lastStageIndex: index,
				error:          err,
			}
		}

		logger.Successf("Test passed.")
	}

	return StageRunnerResult{
		lastStageIndex: len(r.stages) - 1,
		error:          nil,
	}
}

func (r StageRunner) StageCount() int {
	return len(r.stages)
}

// Truncated returns a stageRunner with fewer stages
func (r StageRunner) Truncated(stageSlug string) StageRunner {
	newStages := make([]Stage, 0)
	for _, stage := range r.stages {
		newStages = append(newStages, stage)
		if stage.slug == stageSlug {
			return StageRunner{
				isDebug: r.isDebug,
				stages:  newStages,
			}
		}
	}

	panic(fmt.Sprintf("Stage slug %v not found. Stages: %v", stageSlug, r.stages))
}

// Fuck you, go
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Randomized returns a stage runner that has stages randomized
func (r StageRunner) Randomized() StageRunner {
	return StageRunner{
		isDebug: r.isDebug,
		stages:  shuffleStages(r.stages),
	}
}

func shuffleStages(stages []Stage) []Stage {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	ret := make([]Stage, len(stages))
	perm := r.Perm(len(stages))
	for i, randIndex := range perm {
		ret[i] = stages[randIndex]
	}
	return ret
}

func reportTestError(err error, isDebug bool, logger *customLogger) {
	logger.Errorf("%s", err)
	if isDebug {
		logger.Errorf("Test failed")
	} else {
		logger.Errorf("Test failed " +
			"(try setting 'debug: true' in your codecrafters.yml to see more details)")
	}
}

// Stage is blah
type Stage struct {
	slug    string
	name    string
	runFunc func(executable *Executable, logger *customLogger) error
	logger  *customLogger
}
