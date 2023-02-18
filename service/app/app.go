/* The app package provides constants, properties, and various
 * convenience facilities for the overall application.
 */
package app

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

const (
	Application = "varlog"  // The application name

	// Strings for HTTP response headers
	HdrAttachment = "attachment"
	HdrContentDisposition = "Content-Disposition"
	HdrFilename = "filename"
	HdrInline = "inline"

	LogDebug    = "DEBUG"   // log level: DEBUG
	LogError    = "ERROR"   // log level: ERROR
	LogInfo     = "INFO"    // log level: INFO
	LogWarning  = "WARNING" // log level: WARNING

	ParamContentDisposition = "content-disposition" // name of 'content-disposition' parameter
	ParamCount  = "count"  // Name of the 'count' parameter
	ParamFilter = "filter" // Name of the 'filter' parameter
	ParamName   = "name"   // Name of the 'name' parameter

	// Standard root of the file tree.  This can be updated
	// at program startup.  The rest of the application should
	// use app.Root() to get the correct value.
	pathRoot = "/var/log" // Standard root of file tree

	// The size of chunks read from the file in one Read operation.
	// This can be adjusted for file system performance, typical
	// log file size and host environment.
	productionChunkSize = 64 * 1024

	// Values for the 'list' metadata
	TypeDir  = "dir"
	TypeFile = "file"
)

type Properties struct {
	chunkSize  int    // Chunk size to read from log file
	filterOmit bool   // True if filter text originally had '-'
	filterText string // Filter parameter from request, '-' stripped
	paramContentDisposition string // Desired "Content-Disposition" value
	paramCount int    // Maximum lines to return to client
	paramName  string // Name parameter from request
	root       string // Log directory root.  No trailing slash.
	rootedPath string // full path, e.g., /var/log/dir
}

var properties = Properties{
	root:      pathRoot,
	chunkSize: productionChunkSize,
}

// NewProperties allocates a new Properties object and
// initializes it to the global application properties.
func NewProperties() (p *Properties) {
	p = new(Properties)
	*p = properties
	return p
}

// BasePath returns the last component (base name) of the
// current request's path.  "/var/log/abc" => "abc".
func (p *Properties) BasePath() string {
	return path.Base(p.rootedPath)
}

// ChunkSize provides the chunk size to read from log files.
func (p *Properties) ChunkSize() int {
	return p.chunkSize
}

// Retrieve client parameters from the http request.  Extracts
// the values and updates the properties object that will be used
// for the remainder of this request's processing.
func (props *Properties) ExtractParams(request *http.Request) (err error) {
	if err = request.ParseForm(); err != nil {
		Log(LogError, "%s", err)
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
		case ParamContentDisposition:
			if len(value) == 0 {
				break
			}
			switch value[0] {
			case "", HdrInline, HdrAttachment:
				props.paramContentDisposition = value[0]

			default:
				err = errors.New(
					fmt.Sprintf("Invalid value %s=%q", ParamContentDisposition, value[0]))
				Log(LogWarning, "%s", err.Error())
				return err
			}

		case ParamCount:
			if len(value) == 0 {
				break
			}
			if value[0] == "" {
				props.paramCount = 0
				break
			}
			if props.paramCount, err = strconv.Atoi(value[0]); err != nil {
				err = errors.New(
					fmt.Sprintf("Invalid conversion of param %s=%q, %s",
						ParamCount, value[0], err.Error()))
				Log(LogWarning, "%s", err.Error())
				return err
			}

		case ParamFilter:
			if len(value) == 0 {
				break
			}
			props.filterText = value[0]
			if len(props.filterText) > 0 && props.filterText[0] == '-' {
				props.filterOmit = true
				props.filterText = props.filterText[1:]
			}

		case ParamName:
			if len(value) == 0 {
				break
			}
			err = props.SetParamName(value[0])
			if err != nil {
				err = errors.New(
					fmt.Sprintf("Invalid conversion of param %s=%q, %s",
						ParamName, value[0], err.Error()))
				Log(LogWarning, "%s", err.Error())
				return err
			}

		default:
			// Treat unknown keys as a client error.
			err = errors.New(fmt.Sprintf("Parameter %q invalid", key))
			Log(LogWarning, "%s", err)
			return err
		}
	}
	return nil
}

func (props *Properties) FilterAllowsEntry(name string) bool {
	// An empty filter allows all entries
	if props.filterText == "" {
		return true
	}
	if strings.Contains(name, props.filterText) {
		return !props.filterOmit
	}
	// Filter text is non-empty and did not match.
	return props.filterOmit
}

// FilterOmit indicates whether a match on the filter text
// should omit (true) or include (false) the entry from results.
func (p *Properties) FilterOmit() bool {
	return p.filterOmit
}

func (p *Properties) SetFilterOmit(b bool) {
	p.filterOmit = b
}

func (p *Properties) SetFilterText(s string) {
	p.filterText = s
}

// FilterText provides the value for the 'filter' parameter.
// The value is the empty string if the parameter was not present.
// For matching purposes, a nil/empty value means no filtering
// happens, and all entries qualify for inclusion in results.

// ParamContentDisposition gives the value for any "Content-Disposition"
// header.  The default, empty string, leaves the value up to the server.
// The client can provide an explicit value: "inline" or "attachment".
func (p *Properties) ParamContentDisposition() string {
	return p.paramContentDisposition
}

// ParamCount provides the 'count' parameter's value.  If the
// request did not have the parameter, the value is zero.
// For the /read request, the count caps the line count
// for the results.  This count is applied after filtering.
func (p *Properties) ParamCount() int {
	return p.paramCount
}

// ParamName provides the 'name' parameter's value.  If the
// request did not have the parameter, the string is empty.
func (p *Properties) ParamName() string {
	return p.paramName
}

func (props *Properties) SetParamName(name string) error {
	props.paramName = name

	/* Join the root and the user's path.  The result is cleaned:
	* suppress multiple slashes, process . and .., etc.
	* The result can be the original root (empty name).
	* Otherwise it must have root/ as a prefix.  If not, the
	* input path was trying to go outside the root.
	 */
	p := path.Join(props.root, name)
	if p != props.root && !strings.HasPrefix(p, props.root+"/") {
		err := errors.New(
			fmt.Sprintf("Invalid name parameter (%q)", props.ParamName()))
		Log(LogWarning, "%s", err.Error())
		return err
	}
	props.rootedPath = p
	return nil
}

// Root gives the base directory for all file system operations,
// default is /var/log.  This can be changed for testing.
func (p *Properties) Root() string {
	return p.root
}

// RootedPath gives the full path for an endpoint.
// After an endpoint processes the 'name' parameter relative
// to the Root path, this value provides the result.  For
// example, if name=abc, the rooted path would be /var/log/abc.
func (p *Properties) RootedPath() string {
	return p.rootedPath
}

// SetRootedPath joins Root and ParamName, cleans the resulting
// path and updates the RootedPath value.
// Returns the resulting value.
func (p *Properties) SetRootedPath(s string) {
	p.rootedPath = s
}

// Processes command line arguments, if any.
// The program currently accepts one optional argument, giving
// an alternative rooted path instead of /var/args.
func DoArgs() {
	pwd, _ := os.Getwd()
	Log(LogDebug, "pwd %q", pwd)
	for j, a := range os.Args {
		Log(LogDebug, "Args[%d] %q", j, a)
	}
	if len(os.Args) > 1 {
		Log(LogInfo, "Setting root to %q", os.Args[1])
		properties.root = os.Args[1]
	}
}

// Produces a log entry containing the application name (implicit),
// the log level, and arguments supplied by the caller.
func Log(level string, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	log.Printf("%s %s %s\n",
		Application, level, s)
}

// Provides the active rooted path for the endpoints.
// The default is /var/log, but this can be updated for
// testing and local execution.
// See also: app.SetRoot.
func Root() string {
	return properties.root
}

// Sets the active root directory for the service endpoints.
// Parameters supplied by clients are interpreted relative to
// the active root directory.  This property applies to all
// requests, and thus applies to app, not Properties.
// See also: app.Root.
func SetRoot(root string) {
	properties.root = root
}
