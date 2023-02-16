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
	"bytes"
	"encoding/json"
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

// Metadata for the response.  Note the json package only exports
// public fields.  This uses struct tags to set the key names.
type metadata struct {
	Name string	`json:"name"`	// Item's name, relative to the root
	Type string	`json:"type"`	// Item's type: file or directory
}


// Provides the top-level handler, as called by the HTTP listener.
// Controls overall flow for the endpoint: gather parameters,
// perform the endpoint's actions, write the response.
func Handler(writer http.ResponseWriter, request *http.Request) {
	var props *properties = new(properties)

	app.Log(app.LogInfo, "%q", request.URL)

	// All parameter handling and validation should be done before
	// starting to write the response body (through writer).
	// Otherwise a prelimary write on the response will set the
	// status, and later error handling will not work properly.
	// The response should be unchanged if http.Error() is used at all.

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
	data, err := props.collectMetadata()
	if err != nil {
		http.Error(writer, err.Error(), http.StatusNotFound)
		return
	}
	b, err := json.Marshal(data)
	if err != nil {
		app.Log(app.LogError, "JSON marshal failed: %s", err.Error())
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	// For demonstration purposes, expand and indent the JSON.
	// In production mode, one probably would use the default.
	var out bytes.Buffer
	err = json.Indent(&out, b, "", "  ")
	if err != nil {
		app.Log(app.LogError, "JSON indent failed: %s", err.Error())
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	out.WriteTo(writer)
}


// Generates the response metadata for this request.
// Examines the provided name, classifies it as a file or
// directory, and creates an array for that entity.  A file
// returns information about itself.  A directory gives a list
// of the children.  Any other type of name is an error.
// This function also applies the filter parameter, possibly
// dropping an entry that otherwise would appear in the output.
func (props *properties) collectMetadata() (data []*metadata, err error) {
	fileInfo, err := os.Stat(props.rootedPath)
	if err != nil {
		app.Log(app.LogWarning, "Path %s invalid, %s", props.rootedPath, err.Error())
		return nil, err
	}
	mode := fileInfo.Mode()
	switch {
	case mode.IsDir():
		app.Log(app.LogDebug, "List directory %q", props.rootedPath)
		data, err = props.listDir()

	case mode.IsRegular():
		app.Log(app.LogDebug, "List file %q", props.rootedPath)
		data, err = props.listFile()

	default:
		s := fmt.Sprintf("Special file %q not allowed", props.rootedPath)
		app.Log(app.LogWarning, "%s", s)
		err = errors.New(s)
		return nil, err
	}

	// One last step before returning the results.  All the paths collected
	// in the metadata are full paths, starting with the root directory.
	// We want to remove that root prefix.  The client does not have access
	// to the file system except through the service, and the service should
	// hide anything private.
	root := app.Root() + "/"
	for _, m := range(data) {
		m.stripRootPrefix(root)
	}
	return data, err
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


func (params *properties) filterIncludesEntry(name string) bool {
	// An empty filter allows all entries
	if params.filterText == "" {
		return true
	}
	if strings.Contains(name, params.filterText) {
		return !params.filterOmit
	}
	// Filter text is non-empty and did not match.
	return params.filterOmit
}


// Generate the return metadata for a directory.
func (props *properties) listDir() (data []*metadata, err error) {
	// Need to initialize data away from nil
	data = []*metadata {}
	files, err := os.ReadDir(props.rootedPath)
	if err != nil {
		// This should not happen.  The code already checked the entry
		// is a directory.
		app.Log(app.LogError, "Unable to read directory, %s", err.Error())
		return nil, err
	}
	// Note that os.ReadDir returns a sorted list.  Sorting the resulting
	// metadata array is thus unnecessary.
	for _, file := range(files) {
		if ! props.filterIncludesEntry(file.Name()) {
			continue
		}
		fullPath := path.Join(props.rootedPath, file.Name())
		switch {
		case file.IsDir():
			m := new(metadata)
			m.Name = fullPath
			m.Type = app.TypeDir
			data = append(data, m)

		case file.Type().IsRegular():
			m := new(metadata)
			m.Name = fullPath
			m.Type = app.TypeFile
			data = append(data, m)

		default:
			// Ignore special files
			continue
		}
	}
	return data, nil
}


// Generate the return metadata for a regular file.
// The file itself is the single entry in the output, though
// it might be dropped when the filter is applied.
func (props *properties) listFile() (data []*metadata, err error) {
	// Need to initialize data away from nil
	data = []*metadata {}
	if ! props.filterIncludesEntry(props.name) {
		return data, nil
	}
	m := new(metadata)
	m.Name = props.rootedPath
	m.Type = app.TypeFile
	return append(data, m), nil
}


// Removes the leading root prefix from a metadata name.
// That is, turns "/var/log/dir/f" into "dir/f".
// Go 1.20 (not yet official) has strings.CutPrefix, but this
// uses the .Cut function from current release.
func (m *metadata)stripRootPrefix(root string) {
	_, m.Name, _ = strings.Cut(m.Name, root)
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
	if p != root && ! strings.HasPrefix(p, root + "/") {
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
	return nil
}
