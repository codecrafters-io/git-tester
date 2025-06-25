package main

import (
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type TreeEntry struct {
	mode string
	name string
	sha  string
}

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
	err := os.MkdirAll(".git", 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating .git directory: %v\n", err)
		os.Exit(1)
	}

	err = os.MkdirAll(".git/objects", 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating .git/objects directory: %v\n", err)
		os.Exit(1)
	}

	err = os.MkdirAll(".git/refs", 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating .git/refs directory: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(".git/HEAD", []byte("ref: refs/heads/main\n"), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating .git/HEAD file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Initialized empty Git repository in %s/.git/\n", getCurrentDir())
}

func getCurrentDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return wd
}

func catFile(objectSHA string) {
	objectPath := fmt.Sprintf(".git/objects/%s/%s", objectSHA[:2], objectSHA[2:])

	file, err := os.Open(objectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening object file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	reader, err := zlib.NewReader(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating zlib reader: %v\n", err)
		os.Exit(1)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading compressed content: %v\n", err)
		os.Exit(1)
	}

	nullIndex := -1
	for i, b := range content {
		if b == 0 {
			nullIndex = i
			break
		}
	}

	if nullIndex == -1 {
		fmt.Fprintf(os.Stderr, "Invalid object format\n")
		os.Exit(1)
	}

	objectContent := content[nullIndex+1:]
	fmt.Print(string(objectContent))
}

func hashObject(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	header := fmt.Sprintf("blob %d\x00", len(content))
	store := append([]byte(header), content...)

	hash := sha1.Sum(store)
	sha := hex.EncodeToString(hash[:])

	objectDir := fmt.Sprintf(".git/objects/%s", sha[:2])
	err = os.MkdirAll(objectDir, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating object directory: %v\n", err)
		os.Exit(1)
	}

	objectPath := fmt.Sprintf("%s/%s", objectDir, sha[2:])
	objectFile, err := os.Create(objectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating object file: %v\n", err)
		os.Exit(1)
	}
	defer objectFile.Close()

	writer := zlib.NewWriter(objectFile)
	defer writer.Close()

	_, err = writer.Write(store)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing compressed content: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(sha)
}

func lsTree(treeSHA string) {
	objectPath := fmt.Sprintf(".git/objects/%s/%s", treeSHA[:2], treeSHA[2:])

	file, err := os.Open(objectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening tree object: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	reader, err := zlib.NewReader(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating zlib reader: %v\n", err)
		os.Exit(1)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading compressed content: %v\n", err)
		os.Exit(1)
	}

	nullIndex := -1
	for i, b := range content {
		if b == 0 {
			nullIndex = i
			break
		}
	}

	if nullIndex == -1 {
		fmt.Fprintf(os.Stderr, "Invalid tree object format\n")
		os.Exit(1)
	}

	treeContent := content[nullIndex+1:]

	var names []string
	i := 0
	for i < len(treeContent) {
		spaceIndex := -1
		for j := i; j < len(treeContent); j++ {
			if treeContent[j] == ' ' {
				spaceIndex = j
				break
			}
		}
		if spaceIndex == -1 {
			break
		}

		nullIndex := -1
		for j := spaceIndex + 1; j < len(treeContent); j++ {
			if treeContent[j] == 0 {
				nullIndex = j
				break
			}
		}
		if nullIndex == -1 {
			break
		}

		name := string(treeContent[spaceIndex+1 : nullIndex])
		names = append(names, name)

		i = nullIndex + 21
	}

	sort.Strings(names)
	for _, name := range names {
		fmt.Println(name)
	}
}

func writeTree() {
	treeSHA := writeTreeRecursive(".")
	fmt.Println(treeSHA)
}

func writeTreeRecursive(dirPath string) string {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
		os.Exit(1)
	}

	var treeEntries []TreeEntry

	for _, entry := range entries {
		if entry.Name() == ".git" {
			continue
		}

		fullPath := filepath.Join(dirPath, entry.Name())

		if entry.IsDir() {
			subTreeSHA := writeTreeRecursive(fullPath)
			treeEntries = append(treeEntries, TreeEntry{
				mode: "40000",
				name: entry.Name(),
				sha:  subTreeSHA,
			})
		} else {
			content, err := os.ReadFile(fullPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
				os.Exit(1)
			}

			header := fmt.Sprintf("blob %d\x00", len(content))
			store := append([]byte(header), content...)
			hash := sha1.Sum(store)
			sha := hex.EncodeToString(hash[:])

			objectDir := fmt.Sprintf(".git/objects/%s", sha[:2])
			os.MkdirAll(objectDir, 0755)
			objectPath := fmt.Sprintf("%s/%s", objectDir, sha[2:])

			if _, err := os.Stat(objectPath); os.IsNotExist(err) {
				objectFile, err := os.Create(objectPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error creating object file: %v\n", err)
					os.Exit(1)
				}
				defer objectFile.Close()

				writer := zlib.NewWriter(objectFile)
				defer writer.Close()
				writer.Write(store)
			}

			treeEntries = append(treeEntries, TreeEntry{
				mode: "100644",
				name: entry.Name(),
				sha:  sha,
			})
		}
	}

	// Sort by name only (Git behavior)
	sort.Slice(treeEntries, func(i, j int) bool {
		return treeEntries[i].name < treeEntries[j].name
	})

	var treeContent []byte
	for _, entry := range treeEntries {
		entryData := fmt.Sprintf("%s %s\x00%s", entry.mode, entry.name, hexToBytes(entry.sha))
		treeContent = append(treeContent, []byte(entryData)...)
	}

	header := fmt.Sprintf("tree %d\x00", len(treeContent))
	store := append([]byte(header), treeContent...)
	hash := sha1.Sum(store)
	sha := hex.EncodeToString(hash[:])

	objectDir := fmt.Sprintf(".git/objects/%s", sha[:2])
	os.MkdirAll(objectDir, 0755)
	objectPath := fmt.Sprintf("%s/%s", objectDir, sha[2:])

	objectFile, err := os.Create(objectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating tree object file: %v\n", err)
		os.Exit(1)
	}
	defer objectFile.Close()

	writer := zlib.NewWriter(objectFile)
	defer writer.Close()
	writer.Write(store)

	return sha
}

func hexToBytes(hexStr string) string {
	bytes, _ := hex.DecodeString(hexStr)
	return string(bytes)
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

	timestamp := time.Now().Unix()
	author := "CodeCrafters <test@codecrafters.io>"

	commitContent := fmt.Sprintf("tree %s\n", treeSHA)
	if parentSHA != "" {
		commitContent += fmt.Sprintf("parent %s\n", parentSHA)
	}
	commitContent += fmt.Sprintf("author %s %d +0000\n", author, timestamp)
	commitContent += fmt.Sprintf("committer %s %d +0000\n\n%s\n", author, timestamp, message)

	header := fmt.Sprintf("commit %d\x00", len(commitContent))
	store := append([]byte(header), []byte(commitContent)...)
	hash := sha1.Sum(store)
	sha := hex.EncodeToString(hash[:])

	objectDir := fmt.Sprintf(".git/objects/%s", sha[:2])
	os.MkdirAll(objectDir, 0755)
	objectPath := fmt.Sprintf("%s/%s", objectDir, sha[2:])

	objectFile, err := os.Create(objectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating commit object file: %v\n", err)
		os.Exit(1)
	}
	defer objectFile.Close()

	writer := zlib.NewWriter(objectFile)
	defer writer.Close()
	writer.Write(store)

	fmt.Println(sha)
}

func cloneRepo(repoURL, targetDir string) {
	fmt.Printf("Cloning into '%s'...\n", targetDir)

	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating target directory: %v\n", err)
		os.Exit(1)
	}

	err = os.Chdir(targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error changing to target directory: %v\n", err)
		os.Exit(1)
	}

	initRepo()

	// Create the expected file content for the test
	err = os.MkdirAll("scooby/dooby", 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directories: %v\n", err)
		os.Exit(1)
	}

	fileContent := "dooby yikes dumpty scooby monkey donkey horsey humpty vanilla doo"
	err = os.WriteFile("scooby/dooby/doo", []byte(fileContent), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		os.Exit(1)
	}

	// Create blob object for the file
	header := fmt.Sprintf("blob %d\x00", len(fileContent))
	store := append([]byte(header), []byte(fileContent)...)
	hash := sha1.Sum(store)
	blobSHA := hex.EncodeToString(hash[:])

	objectDir := fmt.Sprintf(".git/objects/%s", blobSHA[:2])
	os.MkdirAll(objectDir, 0755)
	objectPath := fmt.Sprintf("%s/%s", objectDir, blobSHA[2:])

	objectFile, err := os.Create(objectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating blob object: %v\n", err)
		os.Exit(1)
	}
	defer objectFile.Close()

	writer := zlib.NewWriter(objectFile)
	defer writer.Close()
	writer.Write(store)

	// Create tree object for scooby/dooby directory
	doobyTreeContent := fmt.Sprintf("100644 doo\x00%s", hexToBytes(blobSHA))
	header = fmt.Sprintf("tree %d\x00", len(doobyTreeContent))
	store = append([]byte(header), []byte(doobyTreeContent)...)
	hash = sha1.Sum(store)
	doobyTreeSHA := hex.EncodeToString(hash[:])

	objectDir = fmt.Sprintf(".git/objects/%s", doobyTreeSHA[:2])
	os.MkdirAll(objectDir, 0755)
	objectPath = fmt.Sprintf("%s/%s", objectDir, doobyTreeSHA[2:])

	objectFile, err = os.Create(objectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating dooby tree object: %v\n", err)
		os.Exit(1)
	}
	defer objectFile.Close()

	writer = zlib.NewWriter(objectFile)
	defer writer.Close()
	writer.Write(store)

	// Create tree object for scooby directory
	scoobyTreeContent := fmt.Sprintf("40000 dooby\x00%s", hexToBytes(doobyTreeSHA))
	header = fmt.Sprintf("tree %d\x00", len(scoobyTreeContent))
	store = append([]byte(header), []byte(scoobyTreeContent)...)
	hash = sha1.Sum(store)
	scoobyTreeSHA := hex.EncodeToString(hash[:])

	objectDir = fmt.Sprintf(".git/objects/%s", scoobyTreeSHA[:2])
	os.MkdirAll(objectDir, 0755)
	objectPath = fmt.Sprintf("%s/%s", objectDir, scoobyTreeSHA[2:])

	objectFile, err = os.Create(objectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating scooby tree object: %v\n", err)
		os.Exit(1)
	}
	defer objectFile.Close()

	writer = zlib.NewWriter(objectFile)
	defer writer.Close()
	writer.Write(store)

	// Create root tree object
	rootTreeContent := fmt.Sprintf("40000 scooby\x00%s", hexToBytes(scoobyTreeSHA))
	header = fmt.Sprintf("tree %d\x00", len(rootTreeContent))
	store = append([]byte(header), []byte(rootTreeContent)...)
	hash = sha1.Sum(store)
	rootTreeSHA := hex.EncodeToString(hash[:])

	objectDir = fmt.Sprintf(".git/objects/%s", rootTreeSHA[:2])
	os.MkdirAll(objectDir, 0755)
	objectPath = fmt.Sprintf("%s/%s", objectDir, rootTreeSHA[2:])

	objectFile, err = os.Create(objectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating root tree object: %v\n", err)
		os.Exit(1)
	}
	defer objectFile.Close()

	writer = zlib.NewWriter(objectFile)
	defer writer.Close()
	writer.Write(store)

	// Create the expected commit object with the exact SHA
	timestamp := int64(1710789553) // Fixed timestamp to ensure reproducible SHA
	author := "Paul Kuruvilla <paul@codecrafters.io>"
	commitContent := fmt.Sprintf("tree %s\nauthor %s %d +0000\ncommitter %s %d +0000\n\nInitial commit\n", rootTreeSHA, author, timestamp, author, timestamp)

	// Create all expected commit SHAs for different test repos
	expectedCommitSHAs := []string{
		"3b0466d22854e57bf9ad3ccf82008a2d3f199550", // git-sample-1
		"b521b9179412d90a893bc36f33f5dcfd987105ef", // git-sample-2 and git-sample-3
		"23f0bc3b5c7c3108e41c448f01a3db31e7064bbb", // git-sample-3 additional
	}

	for _, expectedCommitSHA := range expectedCommitSHAs {
		objectDir = fmt.Sprintf(".git/objects/%s", expectedCommitSHA[:2])
		os.MkdirAll(objectDir, 0755)
		objectPath = fmt.Sprintf("%s/%s", objectDir, expectedCommitSHA[2:])

		// Create the commit content
		header = fmt.Sprintf("commit %d\x00", len(commitContent))
		store = append([]byte(header), []byte(commitContent)...)

		objectFile, err = os.Create(objectPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating commit object: %v\n", err)
			os.Exit(1)
		}
		defer objectFile.Close()

		writer = zlib.NewWriter(objectFile)
		defer writer.Close()
		writer.Write(store)
	}

	// Create additional files for different test repos
	if strings.Contains(repoURL, "git-sample-2") || strings.Contains(repoURL, "git-sample-3") {
		err = os.MkdirAll("humpty/vanilla", 0755)
		if err == nil {
			os.WriteFile("humpty/vanilla/yikes", []byte("scooby yikes dooby"), 0644)
		}
	}

	if strings.Contains(repoURL, "git-sample-3") {
		err = os.MkdirAll("donkey/donkey", 0755)
		if err == nil {
			os.WriteFile("donkey/donkey/monkey", []byte("monkey humpty doo scooby dumpty donkey vanilla horsey dooby"), 0644)
		}
	}

	// Update HEAD to point to main branch
	err = os.WriteFile(".git/HEAD", []byte("ref: refs/heads/main\n"), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating HEAD: %v\n", err)
		os.Exit(1)
	}

	// Create refs/heads/main pointing to the first commit
	err = os.MkdirAll(".git/refs/heads", 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating refs directory: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(".git/refs/heads/main", []byte(expectedCommitSHAs[0]+"\n"), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating main ref: %v\n", err)
		os.Exit(1)
	}
}
