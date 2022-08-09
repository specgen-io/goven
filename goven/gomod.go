package goven

import (
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"golang.org/x/mod/modfile"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Gomod struct {
	File       *modfile.File
	modulePath string
	ModuleName string

	newModuleName      string
	moduleRenames      map[string]string
	moduleRequireDrops map[string]any
	moduleReplaceDrops map[string]string
}

func OpenModfile(modulePath, modFilename string) (*Gomod, error) {
	file, err := readModfile(filepath.Join(modulePath, modFilename))
	if err != nil {
		return nil, err
	}
	return &Gomod{
		file,
		modulePath,
		file.Module.Mod.Path,
		"",
		map[string]string{},
		map[string]any{},
		map[string]string{},
	}, nil
}

func (mod *Gomod) Rename(name string) {
	mod.newModuleName = name
}

func (mod *Gomod) RenameSubmodule(oldName, newName string) {
	mod.moduleRenames[oldName] = newName
}

func (mod *Gomod) DropRequire(moduleName string) {
	mod.moduleRequireDrops[moduleName] = true
}

func (mod *Gomod) IsRequireDropped(moduleName string) bool {
	_, found := mod.moduleRequireDrops[moduleName]
	return found
}

func (mod *Gomod) DropReplace(moduleName, version string) {
	mod.moduleReplaceDrops[moduleName] = version
}

func (mod *Gomod) Save(modFilename string) error {
	err := renameModules(mod.modulePath, mod.moduleRenames)
	if err != nil {
		return err
	}

	if mod.newModuleName != "" {
		err := renameModules(mod.modulePath, map[string]string{mod.ModuleName: mod.newModuleName})
		if err != nil {
			return err
		}
	}

	for moduleName, version := range mod.moduleReplaceDrops {
		err := mod.File.DropReplace(moduleName, version)
		if err != nil {
			return err
		}
	}

	for require := range mod.moduleRequireDrops {
		err := mod.File.DropRequire(require)
		if err != nil {
			return err
		}
	}

	if mod.newModuleName != "" {
		err := mod.File.AddModuleStmt(mod.newModuleName)
		if err != nil {
			return err
		}
	}

	err = writeModfile(mod.File, filepath.Join(mod.modulePath, modFilename))
	if err != nil {
		return err
	}
	return nil
}

func readModfile(gomodPath string) (*modfile.File, error) {
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

func writeModfile(file *modfile.File, path string) error {
	file.Cleanup()
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

func renameModules(modulePath string, renames map[string]string) error {
	filenamePatterns := []string{"*.go"}
	var err = filepath.Walk(modulePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf(`failed to replace in path "%s" - %s`, path, err.Error())
		}
		if !info.IsDir() {
			_, filename := filepath.Split(path)
			match, err1 := MatchesAny(filenamePatterns, filename)
			if err1 != nil {
				return err1
			}
			if match {
				return renameInFile(path, renames)
			}
		}
		return nil
	})
	return err
}

func renameInFile(path string, renames map[string]string) error {
	fset := token.NewFileSet()
	ast, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf(`failed to parse file "%s" - %s`, path, err.Error())
	}

	for oldName, newName := range renames {
		for _, imp := range ast.Imports {
			if strings.HasPrefix(imp.Path.Value, `"`+oldName) {
				newPathValue := `"` + newName + strings.TrimPrefix(imp.Path.Value, `"`+oldName)
				imp.Path.Value = newPathValue
			}
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf(`failed to code in file "%s" - %s`, path, err.Error())
	}
	err = format.Node(f, fset, ast)
	if err != nil {
		return fmt.Errorf(`failed to write modified imports code to file "%s" - %s`, path, err.Error())
	}
	defer f.Close()
	return nil
}
