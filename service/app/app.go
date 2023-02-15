package app

import (
	"fmt"
	"log"
)

const (
	Application = "varlog"
	LogDebug    = "DEBUG"   	// log level: DEBUG
	LogError    = "ERROR"   	// log level: ERROR
	LogInfo     = "INFO"   	 	// log level: INFO
	LogWarning  = "WARNING"		// log level: WARNING
	pathRoot	= "/var/log"	// Standard root of file tree
)


var properties struct {
	root string
}


func init() {
	properties.root = pathRoot
}


func Log(level string, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	log.Printf("%s %s %s\n",
		Application, level, s)
}


func Root() string {
	return properties.root
}


func SetRoot(root string) {
	properties.root = root
}
