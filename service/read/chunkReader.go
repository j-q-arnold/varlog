package read

import (
	"io"
	"os"
	"varlog/service/app"
)

// This code read a file backwards.  As a log file
// accumulates lines, the most recent appear at the end.
// When viewing lines for diagnostics, the desire is to
// see the most recent lines first.  This code reads
// a log file "backwards", one chunk at a time.
//
// Some edge cases and other considerations.
//
// File reading.
// 1.  Some files could be too big to read into memory as a
//		single blob.  To avoid special cases for large and
//		small files, this code handles all files alike, reading
//		chunks and processing each chunk in turn.
// 2.  Go's bufio package does not allow seeking, so this uses
//		low-level I/O. To avoid unaligned reads, this reads a
//		partial block at the file's end and then backs up to
//		major boundaries for each previous chunk. This uses
//		a moderate power of 2 to agree with file systems.  This
//		chunk size can vary for testing (even to odd values),
//		but use a power of 2 for production.
// 3.  Files can be any size, including zero. The code needs
//		to handle any size file, large or small.
type chunkReader struct {
	file       *os.File
	fileLength int64
	nextOffset int64
	chunkSize  int
	lastError  error
}

// Allocates a new chunkReader and initializes it for use.
// The supplied file will be used for reading, one chunk
// at a time, in reverse order through the file.  The caller
// remains responsible for closing the file.
// The properties give the chunk size the client plans to use.
// Note the caller of the chunk reader
// needs to supply a read buffer to hold chunk data.  That
// buffer should conform to the actual size being used.
// Returns the new chunkReader and an error.
// Returns a nil chunkReader if an error occurs.
func newChunkReader(p *app.Properties, file *os.File) (*chunkReader, error) {
	c := new(chunkReader)
	c.file = file
	c.chunkSize = p.ChunkSize()
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

	case (c.fileLength % int64(c.chunkSize)) == 0:
		// The file is exactly chunked---and not empty.
		// The first read should get the last full chunk.
		c.nextOffset = c.fileLength - int64(c.chunkSize)

	default:
		// The file is not exactly chunked but not empty.
		// Let the first read get the last partial chunk.
		c.nextOffset = c.fileLength - c.fileLength%int64(c.chunkSize)
	}
	return c, nil
}

// peekEOF indicates whether the chunker is at the end,
// and the next read will return EOF.
func (c *chunkReader) peekEOF() bool {
	return c.fileLength == 0 || c.nextOffset < 0
}

// Reads the next chunk from the file, if one exists.
// Important constraint on the supplied slice: b.
// The slice length, len(b), should consistent throughout
// the life of a given chunk reader.  This normally should
// be ChunkSize(), but it can be changed for testing.
// The caller controls the slice capacity (in case the data will
// be extended).
// The return count is the number of bytes actually read.
// A count of zero and error of EOF indicate end of file.
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
	// the actual number of available bytes at the file's tail.
	// No need to adjust the supplied slice length.
	// =When reading the tail chunk, ReadAt can return data and EOF.
	// That EOF needs to be ignored, or the reader stops prematurely.
	count, err = c.file.ReadAt(b, c.nextOffset)
	if count > 0 && err == io.EOF {
		err = nil
	}
	// Subtlety: Always back up the offset by the chunk size.
	// The first pass reads a partial chunk at the end of the file,
	// but we want to back up a full chunk, NOT the read count.
	// This relies on the caller supplying the same slice size for the
	// next read.
	c.nextOffset -= int64(len(b))
	c.lastError = err
	return count, c.lastError
}
