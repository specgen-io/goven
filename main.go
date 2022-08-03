package main

import (
	"flag"
	"fmt"
	"github.com/vsapronov/gotore/localize"
	"os"
)

func main() {
	var gomodPath string
	var outPath string
	var newModuleName string
	var localizedModulesPath string
	var localizedVendored bool

	flag.StringVar(&gomodPath, "module", "./go.mod", "location of go.mod to be vendored")
	flag.StringVar(&outPath, "out", "./out", "path where to put vendored module")
	flag.StringVar(&localizedModulesPath, "vendor", "goven", "internal path where vendored code should be placed")
	flag.StringVar(&newModuleName, "name", "", "name of the module after vendoring")
	flag.BoolVar(&localizedVendored, "required", false, "vendor required modules (needs 'go mod vendor' prior goven, deafult false)")
	flag.Parse()

	err := localize.Localize(gomodPath, outPath, newModuleName, localizedModulesPath, localizedVendored)

	if err != nil {
		println(fmt.Sprintf(`vendoring failed: %s`, err.Error()))
		os.Exit(1)
	}
}
