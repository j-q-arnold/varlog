// Package list provides code for the /list service endpoint.
// A summary of the operation: Given a named file
// or directory, read sub-entries, gather metadata,
// and return ls-type information for those entries.
//
// Parameter 'name=path' provides the partial path, appended
// to the root (default /var/log).  If the resolved path
// is a directory, read directory entries.  If the
// resolved path is a file, give information for that
// file itself.  An empty/missing value lists the root.
//
// Parameter 'filter=text' provides a positive (filter=value)
// or a negative (filter=-value) filter on the entries.  Entries
// must match (or not match) the filter to be included in the
// response.  An empty/missing filter passes all entries.
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

// Metadata for the response.  Note the json package only exports
// public fields.  This uses struct tags to set the key names.
type metadata struct {
	Name string `json:"name"` // Item's name, relative to the root
	Type string `json:"type"` // Item's type: file or directory
}

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
	// The response should be unchanged if http.Error() is used at all.

	err := props.ExtractParams(request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := collectMetadata(props)
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
func collectMetadata(props *app.Properties) (data []*metadata, err error) {
	fileInfo, err := os.Stat(props.RootedPath())
	if err != nil {
		app.Log(app.LogWarning, "Path %s invalid, %s", props.RootedPath(), err.Error())
		return nil, err
	}
	mode := fileInfo.Mode()
	switch {
	case mode.IsDir():
		app.Log(app.LogDebug, "List directory %q", props.RootedPath())
		data, err = listDir(props)

	case mode.IsRegular():
		app.Log(app.LogDebug, "List file %q", props.RootedPath())
		data, err = listFile(props)

	default:
		s := fmt.Sprintf("Special file %q not allowed", props.RootedPath())
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
	for _, m := range data {
		m.stripRootPrefix(root)
	}
	return data, err
}

// Generate the return metadata for a directory.
func listDir(props *app.Properties) (data []*metadata, err error) {
	// Need to initialize data away from nil
	data = []*metadata{}
	files, err := os.ReadDir(props.RootedPath())
	if err != nil {
		// This should not happen.  The code already checked the entry
		// is a directory.
		app.Log(app.LogError, "Unable to read directory, %s", err.Error())
		return nil, err
	}
	// Note that os.ReadDir returns a sorted list.  Sorting the resulting
	// metadata array is thus unnecessary.
	for _, file := range files {
		if !props.FilterAllowsEntry(file.Name()) {
			continue
		}
		fullPath := path.Join(props.RootedPath(), file.Name())
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
func listFile(props *app.Properties) (data []*metadata, err error) {
	// Need to initialize data away from nil
	data = []*metadata{}
	if !props.FilterAllowsEntry(props.ParamName()) {
		return data, nil
	}
	m := new(metadata)
	m.Name = props.RootedPath()
	m.Type = app.TypeFile
	return append(data, m), nil
}

// Removes the leading root prefix from a metadata name.
// That is, turns "/var/log/dir/f" into "dir/f".
// Go 1.20 (not yet official) has strings.CutPrefix, but this
// uses the .Cut function from current release.
func (m *metadata) stripRootPrefix(root string) {
	_, m.Name, _ = strings.Cut(m.Name, root)
}
