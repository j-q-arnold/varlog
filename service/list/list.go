/* Provides code for the /list service endpoint.
 * A summary of the operation: Given a named file
 * or directory, read sub-entries, gather metadata,
 * and return ls-type information for those entries.
 *
 * Parameter 'name=path' provides the partial path, appended
 * to the root (default /var/log).  If the resolved path
 * is a directory, read directory entries.  If the
 * resolved path is a file, give information for that
 * file itself.  An empty/missing value lists the root.
 *
 * Parameter 'filter=text' provides a positive (filter=value)
 * or a negative (filter=-value) filter on the entries.  Entries
 * must match (or not match) the filter to be included in the
 * response.  An empty/missing filter passes all entries.
 */
package list

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"varlog/service/app"
)


// Properties used throughout the package.
// Parameters supplied by the client and values computed
// during request processing that move from one step
// to another.
type properties struct {
	name       string	// name parameter from request
	filterText string	// filter text from request, with '-' stripped
	filterOmit bool		// true if filter text originally had '-'
	rootedPath string	// full path, e.g., /var/log/dir
}


// Provides the top-level handler, as called by the HTTP listener.
// Controls overall flow for the endpoint: gather parameters,
// perform the endpoint's actions, write the response.
func Handler(writer http.ResponseWriter, request *http.Request) {
	app.Log(app.LogInfo, "%q", request.URL)

	// All parameter handling and validation should be done before
	// starting to write the response body (through writer).
	// Otherwise a prelimary write on the response will set the
	// status, and later error handling will not work properly.
	// The response should be "clean" if http.Error() is used at all.

	params, err := extractParams(request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	err = validateParams(params)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusForbidden)
		return
	}
	fmt.Fprintf(writer, "endpoint: list\n")
	fmt.Fprintf(writer, "method %s, full path %q, proto %s\n",
		request.Method, params.rootedPath, request.Proto)
	fmt.Fprintf(writer, "Done\n")
}


func extractParams(request *http.Request) (params *properties, err error) {
	if err = request.ParseForm(); err != nil {
		app.Log(app.LogError, "%s", err)
		return nil, err
	}
	params = new(properties)

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
		fmt.Fprintf(os.Stderr, "param %q = %q\n", key, value)
		switch key {
		case app.ParamFilter:
			if len(value) == 0 {
				break
			}
			params.filterText = value[0]
			if params.filterText[0] == '-' {
				params.filterOmit = true
				params.filterText = params.filterText[1:]
			}

		case app.ParamName:
			if len(value) == 0 {
				break
			}
			params.name = value[0]

		default:
			// Treat unknown keys as a client error.
			err = errors.New(fmt.Sprintf("Parameter %q invalid", key))
			app.Log(app.LogWarning, "%s", err)
			return nil, err
		}
	}
	app.Log(app.LogDebug, "list extract params %+v", params)
	return params, nil
}


func validateParams(params *properties) (err error) {
	// Parameter 'name' validation.

	/* Join the root and the user's path.  The result is cleaned:
	 * suppress multiple slashes, process . and .., etc.
	 * The result can be the original root (empty name).
	 * Otherwise it must have root/ as a prefix.  If not, the
	 * input path was trying to go outside the root.
	 */
	root := app.Root()
	p := path.Join(root, params.name)
	app.Log(app.LogDebug, "joined path %q", p)
	if p != root && ! strings.HasPrefix(p, root + "/") {
		err = errors.New(
			fmt.Sprintf("Invalid name parameter (%q)", params.name))
		app.Log(app.LogWarning, "%s", err.Error())
		return err
	}
	params.rootedPath = p

	// Parameter 'filter' validation: none needed
	// The filter is a simple text string match.
	// If this allowed regex or other matching logic, something would
	// need to go here.
	app.Log(app.LogDebug, "list validate params %+v", params)
	return nil
}
