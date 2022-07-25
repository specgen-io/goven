package main

import (
	"flag"
	"fmt"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func matchesAny(patterns []string, filename string) (bool, error) {
	for _, pattern := range patterns {
		match, err := filepath.Match(pattern, filename)
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

func copyDir(sourcePath, destinationPath string, excludePatterns []string) error {
	var err = filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		var relativePath = strings.Replace(path, sourcePath, "", 1)
		if info.IsDir() {
			return os.Mkdir(filepath.Join(destinationPath, relativePath), 0755)
		} else {
			_, filename := filepath.Split(path)
			match, err1 := matchesAny(excludePatterns, filename)
			if err1 != nil {
				return err1
			}
			if !match {
				var data, err1 = ioutil.ReadFile(filepath.Join(sourcePath, relativePath))
				if err1 != nil {
					return err1
				}
				err1 = ioutil.WriteFile(filepath.Join(destinationPath, relativePath), data, 0777)
				return err1

			}
			return nil
		}
	})
	return err
}

func replace(path string, old, new string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	result := strings.Replace(string(data), old, new, -1)
	err = ioutil.WriteFile(path, []byte(result), 0644)
	return err
}

func replaceInPath(source string, patterns []string, old, new string) error {
	var err error = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			_, filename := filepath.Split(path)
			match, err1 := matchesAny(patterns, filename)
			if err1 != nil {
				return err1
			}
			if match {
				err1 := replace(path, old, new)
				return err1
			}
		}
		return nil
	})
	return err
}

func getLocalModuleName(version module.Version) string {
	newParts := strings.Split(version.Path, string(filepath.Separator))
	localModuleName := newParts[len(newParts)-1]
	return localModuleName
}

func renameModule(modulePath string, old, new string) error {
	err := replaceInPath(modulePath, []string{"*.go"}, old, new)
	if err != nil {
		return err
	}
	return nil
}

func makeLocal(modulePath string, mainModuleName string, replace *modfile.Replace) error {
	localizedSubmoduleName := getLocalModuleName(replace.New)
	sourceModulePath := filepath.Join(modulePath, replace.New.Path)
	localizedModulePath := filepath.Join(modulePath, localizedSubmoduleName)
	err := copyDir(sourceModulePath, localizedModulePath, []string{"go.mod", "go.sum"})
	if err != nil {
		return err
	}
	err = renameModule(modulePath, replace.Old.Path, fmt.Sprintf(`%s/%s`, mainModuleName, localizedSubmoduleName))
	if err != nil {
		return err
	}
	return nil
}

func readGomod(gomodPath string) (*modfile.File, error) {
	buf, err := ioutil.ReadFile(gomodPath)
	if err != nil {
		return nil, err
	}
	gomod, err := modfile.Parse(gomodPath, buf, nil)
	if err != nil {
		return nil, err
	}
	return gomod, nil
}

func writeGomod(file *modfile.File, path string) error {
	data, err := file.Format()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func localize(gomodPath string, newModuleName string) error {
	gomod, err := readGomod(gomodPath)
	if err != nil {
		return err
	}
	mainModuleName := gomod.Module.Mod.Path
	moduleRenames := map[string]string{}
	modulePath, _ := filepath.Split(gomodPath)
	for _, replace := range gomod.Replace {
		localizedSubmoduleName := getLocalModuleName(replace.New)
		sourceModulePath := filepath.Join(modulePath, replace.New.Path)
		localizedModulePath := filepath.Join(modulePath, localizedSubmoduleName)
		err := copyDir(sourceModulePath, localizedModulePath, []string{"go.mod", "go.sum"})
		if err != nil {
			return err
		}
		moduleRenames[replace.Old.Path] = fmt.Sprintf(`%s/%s`, mainModuleName, localizedSubmoduleName)
		moduleName := replace.Old.Path
		moduleVersion := replace.Old.Version
		err = gomod.DropReplace(moduleName, moduleVersion)
		if err != nil {
			return err
		}
		err = gomod.DropRequire(moduleName)
		if err != nil {
			return err
		}
	}

	for old, new := range moduleRenames {
		err = renameModule(modulePath, old, new)
		if err != nil {
			return err
		}
	}

	err = renameModule(modulePath, mainModuleName, newModuleName)
	if err != nil {
		return err
	}

	gomod.AddModuleStmt(newModuleName)
	err = writeGomod(gomod, gomodPath)
	if err != nil {
		return err
	}
	return nil
}

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

	err := localize(gomodPath, newModuleName)

	if err != nil {
		fmt.Println(`Error: %s`, err.Error())
	}
}
