package main

import (
	"flag"
	"fmt"
	"github.com/vsapronov/gotore/goven"
	"os"
)

func main() {
	var gomodPath string
	var outPath string
	var newModuleName string
	var vendoredModulesFolder string
	var vendorRequired bool

	flag.StringVar(&gomodPath, "module", "./go.mod", "location of go.mod to be vendored")
	flag.StringVar(&outPath, "out", "./out", "path where to put vendored module")
	flag.StringVar(&vendoredModulesFolder, "vendor", "goven", "internal path where vendored modules should be placed")
	flag.StringVar(&newModuleName, "name", "", "name of the module after vendoring")
	flag.BoolVar(&vendorRequired, "required", false, "vendor required modules (needs 'go mod vendor' prior goven, deafult false)")
	flag.Parse()

	err := goven.Vendor(gomodPath, outPath, newModuleName, vendoredModulesFolder, vendorRequired)

	if err != nil {
		println(fmt.Sprintf(`vendoring failed: %s`, err.Error()))
		os.Exit(1)
	}
}
