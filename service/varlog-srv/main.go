package main

import (
	"net/http"
	"os"
	"varlog/service/app"
	"varlog/service/list"
	"varlog/service/read"
)

func main() {
	app.Log(app.LogInfo, "starting")

	app.DoArgs()

	// Specify the handler functions for the endpoints.
	http.HandleFunc("/list", list.Handler)
	http.HandleFunc("/read", read.Handler)

	// For the time being, always listen on port 8000.
	// The listener "never" returns.  The documentation says
	// it returns a non-nil error but does not say under what conditions.
	err := http.ListenAndServe("localhost:8000", nil)
	app.Log(app.LogError, "terminating, %s", err)
	os.Exit(1)
}
