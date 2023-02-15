/* Provides code for the /read service endpoint.
 * A summary of the operation: Given a named file,
 * read qualifying lines and present them most-recent
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
	name       string	// name parameter from request
	filterText string	// filter text from request, with '-' stripped
	filterOmit bool		// true if filter text originally had '-'
	count int			// Maximum lines to return to client
	rootedPath string	// full path, e.g., /var/log/dir
}


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
	fmt.Fprintf(writer, "endpoint: read\n")
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
		case app.ParamCount:
			if len(value) == 0 {
				break
			}
			if params.count, err = strconv.Atoi(value[0]); err != nil {
				err = errors.New(
						fmt.Sprintf("Invalid conversion of param %s=%q, %s",
								app.ParamCount, value[0], err.Error()))
				app.Log(app.LogWarning, "%s", err.Error());
				return nil, err
			}

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
	app.Log(app.LogDebug, "read extract params %+v", params)
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
	app.Log(app.LogDebug, "read validate params %+v", params)
	return nil
}