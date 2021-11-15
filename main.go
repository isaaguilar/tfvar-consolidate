package main

import (
	"log"

	"github.com/isaaguilar/tfvar-consolidate/pkg/consolidate"
	flag "github.com/spf13/pflag"
)

func main() {
	var files []string
	var out string
	flag.StringVarP(&out, "out", "o", "", "Path to write consolidated tfvar file too")
	flag.StringSliceVarP(&files, "file", "f", []string{}, "Path and file to tfvar. Specify multiple with commas or use `-f` flag multiple times.")
	flag.Parse()

	err := consolidate.Consolidate(out, files)
	if err != nil {
		log.Fatal(err)
	}
}
