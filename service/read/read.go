/* Provides code for the /read service endpoint.
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
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"varlog/service/app"
)

// Properties used throughout the package.
// Parameters supplied by the client and values computed
// during request processing that move from one step
// to another.
type properties struct {
	name       string // name parameter from request
	filterText string // filter text from request, with '-' stripped
	filterOmit bool   // true if filter text originally had '-'
	count      int    // Maximum lines to return to client
	rootedPath string // full path, e.g., /var/log/dir
	chunkSize  int    // chunk size to read from the log file
}

// Provides the top-level handler, as called by the HTTP listener.
// Controls overall flow for the endpoint: gather parameters,
// perform the endpoint's actions, write the response.
func Handler(writer http.ResponseWriter, request *http.Request) {
	var props *properties = new(properties)

	// For testing purposes, set the chunk size here.
	// Replace 0 (default) with the test value.
	props.chunkSize = 0

	app.Log(app.LogInfo, "%q", request.URL)

	// All parameter handling and validation should be done before
	// starting to write the response body (through writer).
	// Otherwise a prelimary write on the response will set the
	// status, and later error handling will not work properly.
	// The response should be "clean" if http.Error() is used at all.

	err := props.extractParams(request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	err = props.validateParams()
	if err != nil {
		http.Error(writer, err.Error(), http.StatusForbidden)
		return
	}
	props.writeLines(writer)
	fmt.Fprintf(writer, "endpoint: read\n")
	fmt.Fprintf(writer, "method %s, full path %q, proto %s\n",
		request.Method, props.rootedPath, request.Proto)
	fmt.Fprintf(writer, "Done\n")
}

// Retrieve client parameters from the http request.  Extracts
// the values and updates the properties object that will be used
// for the remainder of this request's processing.
func (props *properties) extractParams(request *http.Request) (err error) {
	if err = request.ParseForm(); err != nil {
		app.Log(app.LogError, "%s", err)
		return err
	}

	// ParseForm above generates url.Values, which is a map from
	// a string key to an array of strings.  A given key is allowed
	// to have multiple values, represented in the map's array entries.
	// Note the code below to check the length of map values and select
	// only the first array entry.  As an example:
	//
	// 		url...?a=v1&a=v2
	//
	// generates map["a"] == [ "v1", "v2" ]
	for key, value := range request.Form {
		switch key {
		case app.ParamCount:
			if len(value) == 0 {
				break
			}
			if props.count, err = strconv.Atoi(value[0]); err != nil {
				err = errors.New(
					fmt.Sprintf("Invalid conversion of param %s=%q, %s",
						app.ParamCount, value[0], err.Error()))
				app.Log(app.LogWarning, "%s", err.Error())
				return err
			}

		case app.ParamFilter:
			if len(value) == 0 {
				break
			}
			props.filterText = value[0]
			if len(props.filterText) > 0 && props.filterText[0] == '-' {
				props.filterOmit = true
				props.filterText = props.filterText[1:]
			}

		case app.ParamName:
			if len(value) == 0 {
				break
			}
			props.name = value[0]

		default:
			// Treat unknown keys as a client error.
			err = errors.New(fmt.Sprintf("Parameter %q invalid", key))
			app.Log(app.LogWarning, "%s", err)
			return err
		}
	}
	return nil
}

// Check the client's parameters for validity.  This is mainly
// syntactic checking, without consulting the file system.
func (props *properties) validateParams() (err error) {
	// Parameter 'name' validation.

	/* Join the root and the user's path.  The result is cleaned:
	 * suppress multiple slashes, process . and .., etc.
	 * The result can be the original root (empty name).
	 * Otherwise it must have root/ as a prefix.  If not, the
	 * input path was trying to go outside the root.
	 */
	root := app.Root()
	p := path.Join(root, props.name)
	if p != root && !strings.HasPrefix(p, root+"/") {
		err = errors.New(
			fmt.Sprintf("Invalid name parameter (%q)", props.name))
		app.Log(app.LogWarning, "%s", err.Error())
		return err
	}
	props.rootedPath = p

	// Parameter 'filter' validation: none needed
	// The filter is a simple text string match.
	// If this allowed regex or other matching logic, something would
	// need to go here.

	// Parameter 'count' validation: none needed
	// Any negative/zero count passes all filtered lines.
	// Any positive value sets an upper limit.
	return nil
}

func (props *properties) writeLines(writer http.ResponseWriter) (err error) {
	var r *reverser
	file, err := os.Open(props.rootedPath)
	if err != nil {
		app.Log(app.LogError, "Cannot open %s: %s", props.rootedPath, err.Error())
		return err
	}
	defer file.Close()

	r, err = props.newReverser(file)
	if err != nil {
		app.Log(app.LogError, "Create reverser error for %s: %s", props.rootedPath, err.Error())
		return err
	}
	var total int
	for r.scan() {
		lines := r.lines()
		total += len(lines)
		for j, s := range lines {
			app.Log(app.LogDebug, "reverser %d: '%s'", j, s)
		}
	}
	app.Log(app.LogDebug, "total lines %d", total)
	return r.err()
}
