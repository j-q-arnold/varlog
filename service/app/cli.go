package app

import (
	"flag"
	"fmt"
	"os"
	"path"
)

type CliFlags struct {
	help  bool
	Chunk int
	Port  int
	Root  string
}

var Cli CliFlags

func init() {
	helpUsage := "Request a usage message"
	flag.BoolVar(&Cli.help, "help", false, helpUsage)
	flag.BoolVar(&Cli.help, "?", false, helpUsage)
	flag.IntVar(&Cli.Chunk, "chunk", defaultChunkSize,
		"The byte count for reading file system chunks. "+
			"Zero keeps the default. Otherwise must be positive.")
	flag.IntVar(&Cli.Port, "port", defaultPort,
		"Port on which the service listens for incoming connections. "+
			"Zero keeps the default; otherwise must be positive.")
	flag.StringVar(&Cli.Root, "root", defaultPathRoot,
		"Root directory for all file operations.")
	flag.Usage = usage
}

// Processes command line arguments, if any.
// The program currently accepts one optional argument, giving
// an alternative rooted path instead of /var/args.
func DoCli() {
	parseFlags()
	setProperties()
}

func parseFlags() {
	flag.Parse()

	if Cli.help {
		usage()
		os.Exit(0)
	}

	switch {
	case Cli.Chunk < 0:
		fmt.Fprintf(flag.CommandLine.Output(), "*** Chunk size (%d) cannot be negative.\n", Cli.Chunk)
		os.Exit(1)

	case Cli.Chunk == 0:
		Cli.Chunk = defaultChunkSize
	}
	switch {
	case Cli.Port < 0:
		fmt.Fprintf(flag.CommandLine.Output(), "*** Port (%d) cannot be negative.\n", Cli.Port)
		os.Exit(1)

	case Cli.Port == 0:
		Cli.Port = defaultPort
	}

	if Cli.Root == "" {
		Cli.Root = defaultPathRoot
	}
	Cli.Root = path.Clean(Cli.Root)
	switch Cli.Root {
	case ".", "..", "/":
		fmt.Fprintf(flag.CommandLine.Output(), "*** Invalid root directory (%s)\n", Cli.Root)
		os.Exit(1)
	}
	fileInfo, err := os.Stat(Cli.Root)
	if err != nil || !fileInfo.Mode().IsDir() {
		fmt.Fprintf(flag.CommandLine.Output(), "*** Root (%s) is not a directory.\n", Cli.Root)
		os.Exit(1)
	}
}

func setProperties() {
	properties.chunkSize = Cli.Chunk
	properties.port = Cli.Port
	properties.root = Cli.Root
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}
