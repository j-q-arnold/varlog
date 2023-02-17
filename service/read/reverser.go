package read

import (
	_ "bufio"
	"fmt"
	"net/http"
	"os"
	"varlog/service/app"
)

// These types read a file backwards.  As a log file
// accumulates lines, the most recent appear at the end.
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
// File reading.
// 1.  Some files could be too big to read into memory as a
//		single blob.  To avoid special cases for large and
//		small files, this code handles all files alike, reading
//		blocks at a time, and processing each block in turn.
// 2.  Go's bufio package does not allow seeking, so this uses
//		low-level I/O. To avoid unaligned reads, this reads a
//		partial block at the file's end and then backs up to
//		major boundaries for each previous chunk. This uses
//		bufio.MaxScanTokenSize as the chunk size, having no
//		better value. That current value (64K) is likely to be
//		reasonable.
// 3.  Files can be any size, including zero. The code needs
//		to handle any size file, large or small.
//
// Line scanning.
// 1.  We assume only two conditions about files. A) The first
//		line in a file starts at position 0. B) The last line
//		in a file ends at the last position (whether it has
//		a newline or not).
// 2.  Lines cannot be assumed to be aligned on chunk boundaries.
//		Consequently, the first text in each chunk might have
//		a prefix in the preceding chunk, which will not have been
//		read yet.
// 3.  The possibility of a continuation condition for the first
//		line itself has some edge cases. A) The last line in the
//		preceding chunk ends precisely at the last byte, thus the
//		"suffix" line is actually a complete line. B) The newline
//		for the last line in the preceding chunk occurs at the
//		first byte in the current chunk. This gives the appearance
//		of a zero-length suffix.
// 4.  File formats are not constrained. Lines might be short or
//		long; the code should present what it finds. This does use
//		bufio for scanning, which imposes bufio.MaxTokenScanSize
//		as the maximum token (line) size.  We'll live with that.
type reverser struct {
// todo: fix this
}

func (props *properties) newReverser(file *os.File) (r *reverser, err error) {
	r = new(reverser)
	var c *chunkReader
	c, err = newChunkReader(file)
	if err != nil {
		app.Log(app.LogError, "Nil chunk reader for %s: %s", props.rootedPath, err.Error())
		return nil, err
	}
	// TODO: move action to run
	fmt.Printf("c %#v\n", c)
	var n int
	b := make([]byte, c.ChunkSize())
	total := 0
	for err = nil; err == nil; {
		n, err = c.read(b)
		total += n
		fmt.Printf("n %d, err %v, b.len %d, total %d\n", n, err, len(b), total)
	}
	return r, nil
}

func (r *reverser) run(writer http.ResponseWriter) error {
	return nil
}