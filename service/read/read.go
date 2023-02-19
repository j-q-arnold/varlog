// Package read provides code for the /read service endpoint.
// A summary of the operation: Given a named file,
// read qualifying lines and present the most-recent
// first in the response.
//
// Parameter 'name=path' provides the partial path, appended
// to the root (default /var/log).  The resolved path
// must be a regular file.
//
// Parameter 'filter=text' provides a positive (filter=value)
// or a negative (filter=-value) filter on the lines.  Entries
// must match (or not match) the filter to be included in the
// response.  An empty/missing filter passes all lines.
//
// Parameter 'count=number' caps the number of lines to include
// in the response.  A missing/empty/non-positive value returns
// all lines in the given file.
//
// Parameter 'content-disposition=value' tells whether to include
// a "Content-Disposition" header in the response.  A missing,
// empty, or 'inline' value uses no explicit header, thus streaming
// the result in a browser.  An explicit 'attachment' includes a
// header, which browsers interpret as saving the response in a file.
package read

import (
	"fmt"
	"net/http"
	"os"
	"varlog/service/app"
)

const (
	// Values to decide whether to use inline or attachment results.
	// If the line count is specified and "small", use inline.
	// If the file size is "small", use inline.
	attachFileSize  = 100000
	attachLineCount = 10000
)

// Provides the top-level handler, as called by the HTTP listener.
// Controls overall flow for the endpoint: gather parameters,
// perform the endpoint's actions, write the response.
func Handler(writer http.ResponseWriter, request *http.Request) {
	var props *app.Properties = app.NewProperties()

	app.Log(app.LogInfo, "%q", request.URL)

	// All parameter handling and validation should be done before
	// starting to write the response body (through writer).
	// Otherwise a prelimary write on the response will set the
	// status, and later error handling will not work properly.
	// The response should be "clean" if http.Error() is used at all.

	err := props.ExtractParams(request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	err = writeLines(props, writer)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
	}
}

// selectContentDisposition optionally adds a "Content-Disposition" header to the response.
// If the response is likely to be large, this directs the client to save
// the results in a file instead of displaying directly. If any errors occur,
// they are ignored and no header is written.  Note this does not use the filter
// in the decision.  That obviously affects the result line count, but there's
// no way to know the filter's likely effect.  And the header needs to be
// written before the result's actual line count is known.
// The header to be added:
//
//	Content-Disposition: attachment; filename="name"
func selectContentDisposition(props *app.Properties, writer http.ResponseWriter, file *os.File) {
	switch props.ParamContentDisposition() {
	case app.HdrInline:
		return

	case app.HdrAttachment:
		break

	default:
		if props.ParamCount() > 0 && props.ParamCount() < attachLineCount {
			return
		}
		fileInfo, err := file.Stat()
		if err != nil {
			return
		}
		fileSize := fileInfo.Size()
		if fileSize < attachFileSize {
			return
		}
	}
	s := fmt.Sprintf("%s; %s=%q", app.HdrAttachment, app.HdrFilename, props.BasePath())
	header := writer.Header()
	header.Add(app.HdrContentDisposition, s)
}

func writeLines(props *app.Properties, writer http.ResponseWriter) (err error) {
	var r *reverser
	file, err := os.Open(props.RootedPath())
	if err != nil {
		app.Log(app.LogWarning, "Cannot open %s: %s", props.RootedPath(), err.Error())
		return err
	}
	defer file.Close()

	selectContentDisposition(props, writer, file)

	r, err = newReverser(props, file)
	if err != nil {
		app.Log(app.LogError, "Create reverser error for %s: %s", props.RootedPath(), err.Error())
		return err
	}
	var total int
countLabel:
	for r.scan() {
		lines := r.lines()
		for _, s := range lines {
			if !props.FilterAllowsEntry(s) {
				continue
			}
			fmt.Fprintln(writer, s)
			total++
			if props.ParamCount() > 0 && total >= props.ParamCount() {
				break countLabel
			}
		}
	}
	app.Log(app.LogInfo, "Read %q, %d lines", props.RootedPath(), total)
	return r.err()
}
