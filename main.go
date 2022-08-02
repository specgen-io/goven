package main

import (
	"flag"
	"fmt"
	"github.com/vsapronov/gotore/localize"
	"os"
)

func main() {
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage: gotore <gomod> <module>\n")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println(`  gomod`)
		fmt.Println(`        Path to go.mod file to work on.`)
		fmt.Println(`  module`)
		fmt.Println(`        New module name.`)
	}

	if len(os.Args) < 3 {
		flag.Usage()
		os.Exit(1)
	}

	gomodPath := os.Args[1]
	newModuleName := os.Args[2]

	err := localize.Localize(gomodPath, newModuleName, "goven")

	if err != nil {
		fmt.Println(`Error: %s`, err.Error())
	}
}
