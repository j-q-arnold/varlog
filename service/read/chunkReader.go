package read

import (
	"bufio"
	"io"
	"os"
)

type chunkReader struct {
	file *os.File
	fileLength int64
	nextOffset int64
	chunkSize int
	lastError error
}

// Allocates a new chunkReader and initializes it for use.
// The supplied file will be used for reading, one chunk
// at a time, in reverse order through the file.  The caller
// remains responsible for closing the file.
// Returns a nil chunkReader if an error occurs.
func newChunkReader(file *os.File) (* chunkReader, error) {
	c := new(chunkReader)
	c.file = file
	c.chunkSize = bufio.MaxScanTokenSize
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	c.fileLength = fileInfo.Size()

	// Compute the offset of the next (first) chunk to read.
	switch {
	case c.fileLength == 0:
		// The file is empty.  Do nothing gracefully.
		c.nextOffset = 0

	case (c.fileLength % bufio.MaxScanTokenSize) == 0:
		// The file is exactly chunked---and not empty.
		// The first read should get the last full chunk.
		c.nextOffset = c.fileLength - bufio.MaxScanTokenSize

	default:
		// The file is not exactly chunked but not empty.
		// Let the first read get the last partial chunk.
		c.nextOffset = c.fileLength - c.fileLength % bufio.MaxScanTokenSize
	}
	return c, nil
 }

 func (c *chunkReader) ChunkSize() int {
	return c.chunkSize
 }

// Reads the next chunk from the file, if one exists.
// Important constraint on the supplied slice: b.
// The slice length, len(b), should consistent throughout
// the life of a given chunk reader.  This normally should
// be ChunkSize(), but it can be changed for testing.
// The caller controls the slice capacity (in case the data will
// be extended).
// The return count is the number of bytes actually read.
// An error of EOF indicates end of file.  This might accompany
// the last non-empty read or a final read returning zero bytes.
func (c *chunkReader) read(b []byte) (count int, err error) {
	// Handle special cases first: Nothing to read or EOF.
	// Note the code below sets nextOffset negative after
	// reading the file's offset=0 chunk.
	if c.fileLength == 0 || c.nextOffset < 0 {
		c.lastError = io.EOF
		return 0, io.EOF
	}
	if c.lastError != nil {
		return 0, c.lastError
	}
	// Rely on the caller to set len(b) appropriately.
	// When using ReadAt, we can request a full chunk and get
	// the actual number of available bytes.  No need to adjust
	// the supplied slice length.
	// Read from the ongoing offset into the caller's slice.
	// When reading the offset=last chunk, ReadAt can return data and EOF.
	// That EOF needs to be ignored or the reader stops prematurely.
	count, err = c.file.ReadAt(b, c.nextOffset)
	if count > 0 && err == io.EOF {
		err = nil
	}
	// Subtlety: Always back up the offset by the chunk size.
	// The first pass reads a partial chunk at the end of the file,
	// but we want to back up a full chunk after that, NOT the read count.
	// This relies on the caller supplying the same slice size for the
	// next read.
	c.nextOffset -= int64(len(b))
	c.lastError = err
	return count, c.lastError
 }