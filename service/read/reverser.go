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

// Reverser presents a file backwards, parsed as lines.
// As a log file accumulates lines, the most recent appear at the end.
// When viewing lines for diagnostics, the desire is to
// see the most recent lines first.  This code reads
// a log file "backwards", splits each chunk into lines,
// and then reverses the lines for presentation.
//
// This process has three coordinating activites: reading
// the file in reverse chunks, splitting those chunks into
// lines, reversing the order of the lines, one chunk
// at a time.  By reading the file backwards, each set of
// extracted lines can be reversed, giving the desired order.
//
// Some edge cases and other considerations.
//
// File reading:  See chunkReader.go.
//
// Line scanning.
//  1. We assume only two conditions about files. A) The first
//     line in a file starts at position 0. B) The last line
//     in a file ends at the last position (terminal newline
//     is optional).
//  2. Lines cannot be assumed to align on chunk boundaries.
//     Consequently, the first text in each chunk might have
//     a prefix in the preceding chunk, which will not have been
//     read yet.
//  3. The possibility of a continuation condition for a chunk's
//     first line itself has some edge cases. Details below.
//  4. File formats are not constrained. Lines might be short or
//     long; the code should present what it finds. This uses
//     bufio for scanning, which imposes bufio.MaxTokenScanSize
//     as the maximum token (line) size.  We'll live with that.
type reverser struct {
	props      *app.Properties // The application properties
	chunker    *chunkReader    // Reads file chunks in reverse order
	chunk      []byte          // Bytes read for processing
	lastError  error           // The last error encountered
	lineSuffix []byte          // Handles cross-chunk line splits.  Details below
}

/* Notes about cross-chunk line handling.
 * Chunks are read in reverse order.  This uses numbering for clarity,
 * where chunks appear in natural increasing order: n-1, n, n+1, etc.
 * The chunk reader presents the chunks in order n, n-1, n-2, etc.
 * The first line of chunk n might be a continuation of the last
 * line of chunk n-1.  That potential suffix text from chunk n has
 * edge cases that must be handled.
 * a) The last line in n-1 has a newline in the last byte.  Thus the
 *		last line and the suffix lines represent two lines, not one.
 * b) The first byte of chunk n is a newline. The bufio scanner presents
 *		this as an empty line. This causes ambiguity when reading n-1.
 *		If n-1's last byte is a newline, this condition should give
 *		two lines, not one.
 *		Long story short, an empty suffix must append a newline, not the
 *		empty string from bufio.
 *
 * Summary for handling block n.
 * - If the first line is empty, use "\n" as the suffix.
 * - If the first line is not empty, use it unmodifed as the suffix.
 *
 * Save the resulting suffix for processing chunk n-1. Append the suffix to
 * the n-1 chunk and hand that to bufio for line scanning.
 */

// newReverser allocates a new object and initializes it to read
// the supplied file. Note the reverser uses a chunkReader for low-level
// input. This reads the file backwards with io.ReadAt, which is not
// available from a simple Reader interface.
func newReverser(props *app.Properties, file *os.File) (r *reverser, err error) {
	r = new(reverser)
	r.props = props
	r.chunker, err = newChunkReader(props, file)
	if err != nil {
		app.Log(app.LogError, "Nil chunk reader for %s: %s", props.RootedPath(), err.Error())
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
	// with length zero but capacity to grow with low startup cost.
	lines = make([]string, 0, initialLineCapacity)

	// Create a bytes.Buffer to hold the chunk data read from
	// the file.  This gives a convenient way to let bufio
	// parse the lines from the chunk.
	buffer = bytes.NewBuffer(r.chunk)
	scanner := bufio.NewScanner(buffer)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	// The scanner "should" not raise an error, but it does if the data
	// are not lines (token too long).  For purposes here, just take the
	// lines provided (maybe nothing) and record the error to stop this file.
	// TODO: consider more sophisticated error recovery.
	if err := scanner.Err(); err != nil {
		app.Log(app.LogError, "Scanner error ignored (probably reading non-text): %s,", err.Error())
		r.lastError = err
	}
	r.saveLineSuffix(&lines)

	// Reverse the lines
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
	return lines
}

func (r *reverser) saveLineSuffix(lines *[]string) {
	// If the chunker is done, leave lines[0] alone.
	if r.chunker.peekEOF() {
		r.lineSuffix = []byte{}
		return
	}
	// Save the first line as the suffix for the next chunk.
	// The entries in lines are new strings from scanner.Text()
	// and safe to use later.  Copy unnecessary.
	if len(*lines) == 0 {
		r.lineSuffix = []byte{}
	} else {
		s := (*lines)[0]
		if s == "" {
			s = "\n"
		}
		r.lineSuffix = []byte(s)
		(*lines) = (*lines)[1:]
	}
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

	// Make a chunk buffer for the low-level chunker to use.
	// After the file has been read into the buffer, we append
	// the reserved line suffix for split-line handling.  That
	// aggregate buffer is then used for parsing into lines.
	r.chunk = make([]byte, r.props.ChunkSize(), r.props.ChunkSize()+len(r.lineSuffix))
	n, r.lastError = r.chunker.read(r.chunk)
	r.chunk = r.chunk[0:n]
	if len(r.lineSuffix) > 0 {
		r.chunk = append(r.chunk, r.lineSuffix...)
	}
	return r.lastError == nil
}
