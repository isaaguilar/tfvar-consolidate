package main

import (
	"log"

	"github.com/isaaguilar/tfvar-consolidate/pkg/consolidate"
	flag "github.com/spf13/pflag"
)

func main() {
	var files []string
	var out string
	var useEnvs bool
	var backend string
	flag.StringVarP(&out, "out", "o", "", "Path to write consolidated tfvar file too")
	flag.StringSliceVarP(&files, "var-file", "f", []string{}, "The tfvar files to consolidate. JSON tfvar files must have .json extensions. Specify multiple with commas or use `-f` flag multiple times.")
	flag.BoolVarP(&useEnvs, "use-envs", "e", false, "Take 'TF_VAR_' environment variables into account")
	flag.StringVarP(&backend, "backend", "b", "", "A backend config (.tf) file. The values of the backend config will be extracted to a .conf file.")
	flag.Parse()

	err := consolidate.Consolidate(out, files, useEnvs, backend)
	if err != nil {
		log.Fatal(err)
	}
}
