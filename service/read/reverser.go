package read

import (
	"bufio"
	"bytes"
	_ "fmt"
	"io"
	"os"
	"varlog/service/app"
)

const (
	// A "guess" at a sensible capacity for the slice used to hold
	// lines from a single file chunk.  This should be adjusted if
	// the value causes considerable internal slice reallocation.
	// Related to the chunk size: bigger chunks would have more lines
	// and warrant a larger initial value here.
	initialLineCapacity = 1000
)

// Reverser presents a file backwards, one line at a time.
// As a log file accumulates lines, the most recent appear at the end.
// When viewing lines for diagnostics, the desire is to
// see the most recent lines first.  This code reads
// a log file "backwards", splits each block into lines,
// and then reverses the lines for presentation.
//
// This process has three coordinating activites: reading
// the file in reverse chunks, splitting those chunks into
// lines, reversing the order of the lines, one chunk
// at a time.  By reading the file backwards, each set of
// extracted lines, when reversed, will be in the desired order.
//
// Some edge cases and other considerations.
//
// File reading:  See chunkReader.
//
// Line scanning.
// 1.  We assume only two conditions about files. A) The first
//		line in a file starts at position 0. B) The last line
//		in a file ends at the last position (terminal newline
//		is optional).
// 2.  Lines cannot be assumed to align on chunk boundaries.
//		Consequently, the first text in each chunk might have
//		a prefix in the preceding chunk, which will not have been
//		read yet.
// 3.  The possibility of a continuation condition for the first
//		chunk line itself has some edge cases. A) The last line in the
//		preceding chunk ends precisely at the last byte, thus the
//		"suffix" line is actually a complete line. B) The newline
//		for the last line in the preceding chunk occurs at the
//		first byte in the current chunk. This gives the appearance
//		of a zero-length suffix.
// 4.  File formats are not constrained. Lines might be short or
//		long; the code should present what it finds. This uses
//		bufio for scanning, which imposes bufio.MaxTokenScanSize
//		as the maximum token (line) size.  We'll live with that.
type reverser struct {
	chunker *chunkReader
	chunkSize int
	chunk []byte
	lastError error
}

func (props *properties) newReverser(file *os.File) (r *reverser, err error) {
	r = new(reverser)
	r.chunker, r.chunkSize, err = newChunkReader(file, props.chunkSize)
	if err != nil {
		app.Log(app.LogError, "Nil chunk reader for %s: %s", props.rootedPath, err.Error())
		return nil, err
	}
	return r, nil
}

// Returns the most recent error for the reverser,
// nil if no error has occurred.  Note that internal io.EOF is a
// normal condition and presents as nil externally.  The scanner
// simply stops in that situation.
func (r *reverser) err() error {
	if r.lastError == io.EOF {
		return nil
	}
	return r.lastError
}

// Extracts lines from the last chunk read from the file.
func (r *reverser) lines() []string {
	var lines []string
	var buffer *bytes.Buffer

	// Allocate a slice for the lines in the chunk.  It starts
	// with zero current length but capacity to grow with low startup cost.
	lines = make([]string, 0, initialLineCapacity)

	// Create a bytes.Buffer to hold the chunk data read from
	// the file.  This gives a convenient way to let bufio
	// parse the lines from the chunk.
	buffer = bytes.NewBuffer(r.chunk)
	scanner := bufio.NewScanner(buffer)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	// Reverse the lines
	for i, j := 0, len(lines) - 1; i < j; i, j = i + 1, j - 1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
	return lines
}

// Advances the reverser to the next chunk of the file being read,
// which will then be available through lines().
// Returns false when the scan should stop, either exhausting the data
// or finding an error.  When scan() returns false, the caller can use
// err() to retrieve the final error (nil on io.EOF).
// Returns true if the reverser has data for the caller
// to process---and by implication should continue calling scan().
func (r *reverser) scan() bool {
	var n int
	if r.lastError != nil {
		return false
	}
	r.chunk = make([]byte, r.chunkSize)
	n, r.lastError = r.chunker.read(r.chunk)
	r.chunk = r.chunk[0:n]
	return r.lastError == nil
}
