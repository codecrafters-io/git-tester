package main

import (
	"math/rand"
	"strings"
	"time"
)

func runGitCmdUnsafe(workingDir string, args ...string) (ExecutableResult, error) {
	executable := NewExecutable("git")
	executable.WorkingDir = workingDir
	return executable.Run(args...)
}

func runGitCmd(workingDir string, args ...string) string {
	executable := NewExecutable("git")
	executable.WorkingDir = workingDir
	result, err := executable.Run(args...)
	if err != nil {
		panic(err)
	}
	return string(result.Stdout)
}

var randomWords = []string{
	"humpty",
	"dumpty",
	"horsey",
	"donkey",
	"yikes",
	"monkey",
	"doo",
	"scooby",
	"dooby",
	"vanilla",
}

func randomString() string {
	rand.Seed(time.Now().UnixNano())

	return strings.Join(
		[]string{
			randomWords[rand.Intn(10)],
			randomWords[rand.Intn(10)],
			randomWords[rand.Intn(10)],
			randomWords[rand.Intn(10)],
			randomWords[rand.Intn(10)],
			randomWords[rand.Intn(10)],
		},
		" ",
	)
}

func randomStringShort() string {
	rand.Seed(time.Now().UnixNano())
	return randomWords[rand.Intn(10)]
}
