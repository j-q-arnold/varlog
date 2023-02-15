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

const (
	paramFilter = "filter"	// Name of the filter parameter
	paramName   = "name"	// Name of the name parameter
)

type Params struct {
	Name       string	// name parameter from request
	FilterText string	// filter text from request, with '-' stripped
	FilterOmit bool		// true if filter text originally had '-'
	rootedPath string	// full path: root / name
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
	fmt.Fprintf(writer, "endpoint: list\n")
	fmt.Fprintf(writer, "method %s, URL %q, proto %s\n",
		request.Method, request.URL, request.Proto)
	fmt.Fprintf(writer, "Done\n")
}


func extractParams(request *http.Request) (params *Params, err error) {
	if err = request.ParseForm(); err != nil {
		app.Log(app.LogError, "%s", err)
		return nil, err
	}
	params = new(Params)
	for key, value := range request.Form {
		fmt.Fprintf(os.Stderr, "param %q = %q\n", key, value)
		switch key {
		case paramFilter:
			if len(value) == 0 {
				break
			}
			params.FilterText = value[0]
			if params.FilterText[0] == '-' {
				params.FilterOmit = true
				params.FilterText = params.FilterText[1:]
			}

		case paramName:
			if len(value) == 0 {
				break
			}
			params.Name = value[0]

		default:
			err = errors.New(fmt.Sprintf("Parameter %q invalid", key))
			app.Log(app.LogWarning, "%s", err)
			return nil, err
		}
	}
	app.Log(app.LogDebug, "list params %+v", params)
	return params, nil
}


func validateParams(params *Params) (err error) {
	// Validate NAME

	/* Join the root and the user's path.  The result is cleaned:
	 * suppress multiple slashes, process . and .., etc.
	 * The result can be the original root (empty name).
	 * Otherwise it must have root/ as a prefix.  If not, the
	 * input path was trying to go outside the root.
	 */
	root := app.Root()
	p := path.Join(root, params.Name)
	app.Log(app.LogDebug, "joined path %q", p)
	if p != root && ! strings.HasPrefix(p, root + "/") {
		err = errors.New(
			fmt.Sprintf("Invalid name parameter (%q)", params.Name))
		app.Log(app.LogWarning, "%s", err.Error())
		return err
	}

	// Validate FILTER, which is either "" or not.
	// No validation needed, because it is a simple text string match.
	// If this allowed regex or other matching logic, something would
	// need to go here.

	return nil
}
