package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"varlog/service/app"
)

var apps []string = []string {
	"aaaaa",
	"bbbbb",
	"ccccc",
	"ddddd",
	"eeeee",
	"fffff",
	"ggggg",
	"hhhhh",
	"iiiii",
	"jjjjj",
}

var levels []string = []string {
	app.LogDebug,
	app.LogInfo,
	app.LogWarning,
	app.LogError,
}

func main() {
	var count int = 20
	var err error

	if len(os.Args) > 1 {
		if count, err = strconv.Atoi(os.Args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "*** Expected argument (%s) to be a number\n",
					os.Args[1])
			os.Exit(1)
		}
		if count <= 0 {
			fmt.Fprintf(os.Stderr, "*** Expected line count (%d) to be positive\n", count)
			os.Exit(1)
		}
	}

	for j := 0; j < count; j++ {
		log.Printf("%s %10d %7s abcde fghij klmno pqrst uvwxy\n",
			apps[j % len(apps)], j, levels[j % len(levels)])
	}
	os.Exit(0)
}
