package main

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
)

func cloneRepoNew(repoURL, targetDir string) {
	fmt.Printf("Cloning into '%s'...\n", targetDir)

	// Use go-git to perform the clone
	_, err := git.PlainClone(targetDir, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error cloning repository: %v\n", err)
		os.Exit(1)
	}
}
