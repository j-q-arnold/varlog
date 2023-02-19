package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

func main() {
	var count int = 20
	var err error
	var w = bufio.NewWriter(os.Stdout)

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

	var letter = byte('a')
	for j := 0; j < count; {
		w.WriteByte(letter)
		if j++; j % 50 == 0 {
			switch letter {
			case 'z': letter = 'A'
			case 'Z': letter = 'a'
			default: letter++
			}
		}
	}
	w.Flush()
	os.Exit(0)
}
