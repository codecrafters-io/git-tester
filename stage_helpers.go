package main

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
