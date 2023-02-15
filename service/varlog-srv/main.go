package main

import (
	"net/http"
	"os"
	"varlog/service/app"
	"varlog/service/list"
	"varlog/service/read"
)

const (
	Application = "varlog"
)

func main() {
	app.Log(app.LogInfo, "starting")
	http.HandleFunc("/list", list.Handler)
	http.HandleFunc("/read", read.Handler)
	err := http.ListenAndServe("localhost:8000", nil)
	app.Log(app.LogError, "terminating, %s", err)
	os.Exit(1)
}
