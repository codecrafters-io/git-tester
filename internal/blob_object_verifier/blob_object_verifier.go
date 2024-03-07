package blob_object_verifier

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/codecrafters-io/tester-utils/bytes_diff_visualizer"
	"github.com/codecrafters-io/tester-utils/logger"
)

type BlobObjectVerifier struct {
	RawContents []byte
}

func (b *BlobObjectVerifier) ExpectedSha() string {
	sha1Hash := sha1.New()
	sha1Hash.Write(b.ExpectedDecompressedFileContents())

	return fmt.Sprintf("%x", sha1Hash.Sum(nil))
}

func (b *BlobObjectVerifier) ExpectedDecompressedFileContents() []byte {
	return []byte(fmt.Sprintf("blob %d\x00%s", len(b.RawContents), b.RawContents))
}

func (b *BlobObjectVerifier) PrintFriendlyDiff(logger *logger.Logger, actualDecompressedFileContests []byte) {
	lines := bytes_diff_visualizer.VisualizeByteDiff(actualDecompressedFileContests, b.ExpectedDecompressedFileContents())
	logger.Errorf("Git object file doesn't match official Git implementation. Diff after zlib decompression:")
	logger.Errorf("")

	for _, line := range lines {
		logger.Plainln(line)
	}

	logger.Errorf("")
}

func (b *BlobObjectVerifier) VerifyFileContents(logger *logger.Logger, repoDir string, actualSha string) error {
	actualFileRelativePath := path.Join(".git", "objects", actualSha[:2], actualSha[2:])
	actualFilePath := path.Join(repoDir, actualFileRelativePath)

	actualFileContents, err := os.ReadFile(actualFilePath)
	if err != nil {
		return fmt.Errorf("Did not find file at %q", actualFileRelativePath)
	}

	compressedActualFileReader, err := zlib.NewReader(bytes.NewReader(actualFileContents))
	if err != nil {
		return fmt.Errorf("The file at %q is not Zlib-compressed", actualFileRelativePath)
	}

	actualDecompressedFileContests, err := io.ReadAll(compressedActualFileReader)
	if err != nil {
		return fmt.Errorf("The file at %q is not Zlib-compressed", actualFileRelativePath)
	}

	if !bytes.Equal(actualDecompressedFileContests, b.ExpectedDecompressedFileContents()) {
		b.PrintFriendlyDiff(logger, actualDecompressedFileContests)
		return fmt.Errorf("File at %q does not match official Git implementation", actualFileRelativePath)
	}

	return nil
}

func emptyZlibCompressedBytes() []byte {
	var in bytes.Buffer
	b := []byte("")
	w := zlib.NewWriter(&in)
	w.Write(b)
	w.Close()

	return in.Bytes()
}
