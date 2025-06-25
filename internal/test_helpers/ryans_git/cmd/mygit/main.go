package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		initRepo()
	case "cat-file":
		if len(os.Args) < 4 || os.Args[2] != "-p" {
			fmt.Fprintf(os.Stderr, "usage: mygit cat-file -p <object>\n")
			os.Exit(1)
		}
		catFile(os.Args[3])
	case "hash-object":
		if len(os.Args) < 4 || os.Args[2] != "-w" {
			fmt.Fprintf(os.Stderr, "usage: mygit hash-object -w <file>\n")
			os.Exit(1)
		}
		hashObject(os.Args[3])
	case "ls-tree":
		if len(os.Args) < 4 || os.Args[2] != "--name-only" {
			fmt.Fprintf(os.Stderr, "usage: mygit ls-tree --name-only <tree-sha>\n")
			os.Exit(1)
		}
		lsTree(os.Args[3])
	case "write-tree":
		writeTree()
	case "commit-tree":
		commitTree(os.Args[2:])
	case "clone":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "usage: mygit clone <repository> <directory>\n")
			os.Exit(1)
		}
		cloneRepo(os.Args[2], os.Args[3])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

func initRepo() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	_, err = git.PlainInit(wd, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing repository: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Initialized empty Git repository in %s/.git/\n", wd)
}

func catFile(objectSHA string) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening repository: %v\n", err)
		os.Exit(1)
	}

	hash := plumbing.NewHash(objectSHA)
	obj, err := repo.Storer.EncodedObject(plumbing.AnyObject, hash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading object: %v\n", err)
		os.Exit(1)
	}

	reader, err := obj.Reader()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting object reader: %v\n", err)
		os.Exit(1)
	}
	defer reader.Close()

	content := make([]byte, obj.Size())
	_, err = reader.Read(content)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading object content: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(string(content))
}

func hashObject(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	repo, err := git.PlainOpen(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening repository: %v\n", err)
		os.Exit(1)
	}

	obj := repo.Storer.NewEncodedObject()
	obj.SetType(plumbing.BlobObject)
	obj.SetSize(int64(len(content)))

	writer, err := obj.Writer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting object writer: %v\n", err)
		os.Exit(1)
	}

	_, err = writer.Write(content)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing object content: %v\n", err)
		os.Exit(1)
	}
	writer.Close()

	hash, err := repo.Storer.SetEncodedObject(obj)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error storing object: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(hash.String())
}

func lsTree(treeSHA string) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening repository: %v\n", err)
		os.Exit(1)
	}

	hash := plumbing.NewHash(treeSHA)
	tree, err := repo.TreeObject(hash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading tree object: %v\n", err)
		os.Exit(1)
	}

	var names []string
	for _, entry := range tree.Entries {
		names = append(names, entry.Name)
	}

	sort.Strings(names)
	for _, name := range names {
		fmt.Println(name)
	}
}

func writeTree() {
	repo, err := git.PlainOpen(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening repository: %v\n", err)
		os.Exit(1)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting worktree: %v\n", err)
		os.Exit(1)
	}

	// Add all files to the index
	err = worktree.AddGlob(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding files: %v\n", err)
		os.Exit(1)
	}

	// Create a tree from the worktree status
	// status, err := worktree.Status()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting worktree status: %v\n", err)
		os.Exit(1)
	}

	treeHash, err := writeTreeRecursive(repo, ".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing tree: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(treeHash.String())
}

func writeTreeRecursive(repo *git.Repository, dirPath string) (plumbing.Hash, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return plumbing.Hash{}, err
	}

	var treeEntries []object.TreeEntry

	for _, entry := range entries {
		if entry.Name() == ".git" {
			continue
		}

		fullPath := dirPath + "/" + entry.Name()
		if dirPath == "." {
			fullPath = entry.Name()
		}

		if entry.IsDir() {
			// Recursively create subtree
			subTreeHash, err := writeTreeRecursive(repo, fullPath)
			if err != nil {
				continue
			}

			treeEntries = append(treeEntries, object.TreeEntry{
				Name: entry.Name(),
				Mode: 0040000, // Directory mode
				Hash: subTreeHash,
			})
		} else {
			// Create blob for file
			content, err := os.ReadFile(fullPath)
			if err != nil {
				continue
			}

			obj := repo.Storer.NewEncodedObject()
			obj.SetType(plumbing.BlobObject)
			obj.SetSize(int64(len(content)))

			writer, err := obj.Writer()
			if err != nil {
				continue
			}
			writer.Write(content)
			writer.Close()

			hash, err := repo.Storer.SetEncodedObject(obj)
			if err != nil {
				continue
			}

			treeEntries = append(treeEntries, object.TreeEntry{
				Name: entry.Name(),
				Mode: 0100644, // Regular file mode
				Hash: hash,
			})
		}
	}

	// Sort entries by name (Git requirement)
	sort.Slice(treeEntries, func(i, j int) bool {
		return treeEntries[i].Name < treeEntries[j].Name
	})

	// Create tree object
	tree := &object.Tree{Entries: treeEntries}
	treeObj := repo.Storer.NewEncodedObject()
	err = tree.Encode(treeObj)
	if err != nil {
		return plumbing.Hash{}, err
	}

	treeHash, err := repo.Storer.SetEncodedObject(treeObj)
	if err != nil {
		return plumbing.Hash{}, err
	}

	return treeHash, nil
}

func commitTree(args []string) {
	if len(args) < 4 {
		fmt.Fprintf(os.Stderr, "usage: mygit commit-tree <tree-sha> -p <parent-sha> -m <message>\n")
		os.Exit(1)
	}

	treeSHA := args[0]
	var parentSHA, message string

	for i := 1; i < len(args); i++ {
		if args[i] == "-p" && i+1 < len(args) {
			parentSHA = args[i+1]
			i++
		} else if args[i] == "-m" && i+1 < len(args) {
			message = args[i+1]
			i++
		}
	}

	repo, err := git.PlainOpen(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening repository: %v\n", err)
		os.Exit(1)
	}

	treeHash := plumbing.NewHash(treeSHA)

	commit := &object.Commit{
		TreeHash: treeHash,
		Message:  message + "\n",
		Author: object.Signature{
			Name:  "CodeCrafters-Bot",
			Email: "hello@codecrafters.io",
			When:  time.Now(),
		},
		Committer: object.Signature{
			Name:  "CodeCrafters-Bot",
			Email: "hello@codecrafters.io",
			When:  time.Now(),
		},
	}

	if parentSHA != "" {
		parentHash := plumbing.NewHash(parentSHA)
		commit.ParentHashes = []plumbing.Hash{parentHash}
	}

	obj := repo.Storer.NewEncodedObject()
	err = commit.Encode(obj)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding commit: %v\n", err)
		os.Exit(1)
	}

	hash, err := repo.Storer.SetEncodedObject(obj)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error storing commit: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(hash.String())
}

func cloneRepo(repoURL, targetDir string) {
	fmt.Printf("Cloning into '%s'...\n", targetDir)

	_, err := git.PlainClone(targetDir, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error cloning repository: %v\n", err)
		os.Exit(1)
	}
}
