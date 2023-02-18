/* Package read provides code for the /read service endpoint.
 * A summary of the operation: Given a named file,
 * read qualifying lines and present the most-recent
 * first in the response.
 *
 * Parameter 'name=path' provides the partial path, appended
 * to the root (default /var/log).  The resolved path
 * must be a regular file.
 *
 * Parameter 'filter=text' provides a positive (filter=value)
 * or a negative (filter=-value) filter on the lines.  Entries
 * must match (or not match) the filter to be included in the
 * response.  An empty/missing filter passes all lines.
 *
 * Parameter 'count=number' caps the number of lines to include
 * in the response.  A missing/empty/non-positive value returns
 * all lines in the given file.
 */
package read

import (
	"fmt"
	"net/http"
	"os"
	"varlog/service/app"
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
	writeLines(props, writer)
	fmt.Fprintf(writer, "endpoint: read\n")
	fmt.Fprintf(writer, "method %s, full path %q, proto %s\n",
		request.Method, props.RootedPath(), request.Proto)
	fmt.Fprintf(writer, "Done\n")
}

func writeLines(props *app.Properties, writer http.ResponseWriter) (err error) {
	var r *reverser
	file, err := os.Open(props.RootedPath())
	if err != nil {
		app.Log(app.LogError, "Cannot open %s: %s", props.RootedPath(), err.Error())
		return err
	}
	defer file.Close()

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
	app.Log(app.LogDebug, "total lines %d", total)
	return r.err()
}
