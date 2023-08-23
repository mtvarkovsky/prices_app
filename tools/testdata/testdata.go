package main

import (
	"log"
	"os"
	"prices/pkg/testutils"
	"strconv"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("usage: testdata {number_of_lines} {output_dir}")
	}
	lines, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil {
		log.Fatalf("{number_of_lines} must be int")
	}
	testutils.GenerateTestData(int(lines), os.Args[2])
}
