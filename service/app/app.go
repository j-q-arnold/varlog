/* The app package provides constants, properties, and various
 * convenience facilities for the overall application.
 */
package app

import (
	"fmt"
	"log"
	"os"
)

const (
	Application = "varlog"		// The application name
	LogDebug    = "DEBUG"   	// log level: DEBUG
	LogError    = "ERROR"   	// log level: ERROR
	LogInfo     = "INFO"   	 	// log level: INFO
	LogWarning  = "WARNING"		// log level: WARNING

	ParamCount = "count"		// Name of the 'count' parameter
	ParamFilter = "filter"		// Name of the 'filter' parameter
	ParamName = "name"			// Name of the 'name' parameter

	// Standard root of the file tree.  This can be updated
	// at program startup.  The rest of the application should
	// use app.Root() to get the correct value.
	pathRoot	= "/var/log"	// Standard root of file tree

	// Values for the 'list' metadata
	TypeDir = "dir"
	TypeFile = "file"
)


var properties struct {
	root string		// see app.Root()
}


func init() {
	properties.root = pathRoot
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
// the active root directory.
// See also: app.Root.
func SetRoot(root string) {
	properties.root = root
}
