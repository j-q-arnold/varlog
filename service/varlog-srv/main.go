// Package main implements a log server.  For usage:
//
//	$ varlog-srv -help
//
// The primary points of interest:
//   - It serves files from /var/log.  Under the /read endpoint,
//     files are read backwards, presenting the newest lines first.
//   - Communicates HTTP, so it can be exercised with a browser.
//   - It provides two endpoints: /list and /read.  List generates a
//     list of files and directories under a given path.  Read
//     opens a file (only), reads lines in reverse order, and
//     sends selected lines in the response.
//   - Both /list and /read support filtering, giving a
//     text string that a line must contain to qualify for the output.
//     The filter also can be negative, filter=-text, to omit lines
//     that contain the given text.
//
// See the README for full details.
package main

import (
	"fmt"
	"net/http"
	"os"
	"varlog/service/app"
	"varlog/service/list"
	"varlog/service/read"
)

func main() {
	// Process command line flags and arguments.
	app.DoCli()

	// Specify the handler functions for the endpoints.
	http.HandleFunc("/list", list.Handler)
	http.HandleFunc("/read", read.Handler)

	// The listener "never" returns.  The documentation says
	// it returns a non-nil error but does not say under what conditions.
	props := app.NewProperties()
	s := fmt.Sprintf("localhost:%d", props.Port())
	app.Log(app.LogInfo, "starting on %s, root %q", s, props.Root())

	err := http.ListenAndServe(s, nil)
	app.Log(app.LogError, "terminating, %s", err)
	os.Exit(1)
}
