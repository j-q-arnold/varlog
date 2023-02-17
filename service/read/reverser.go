package read

import (
	_ "bufio"
	"fmt"
	"os"
	"varlog/service/app"
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
	// TODO: This will need to be a slice of strings
	slice []byte
	lastError error
}

func (props *properties) newReverser(file *os.File) (r *reverser, err error) {
	r = new(reverser)
	r.chunker, r.chunkSize, err = newChunkReader(file, 0)
	if err != nil {
		app.Log(app.LogError, "Nil chunk reader for %s: %s", props.rootedPath, err.Error())
		return nil, err
	}
	fmt.Printf("new reverser, chunker %#v\n", r.chunker)
	return r, nil
}

func (r *reverser) err() error {
	return r.lastError
}

func (r *reverser) scan() bool {
	var n int
	if r.lastError != nil {
		return false
	}
	r.slice = make([]byte, r.chunkSize)
	n, r.lastError = r.chunker.read(r.slice)
	r.slice = r.slice[0:n]
	fmt.Printf("n %d, err %v, b.len %d\n", n, r.lastError, len(r.slice))
	return r.lastError == nil
}

func (r *reverser) text() []byte {
	return r.slice
}
