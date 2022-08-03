package goven

import (
	"fmt"
	"golang.org/x/mod/modfile"
	"io/ioutil"
	"path/filepath"
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
	mod.moduleRenames[oldName] = fmt.Sprintf(`%s/%s`, mod.ModuleName, newName)
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
		err := renameModule(mod.modulePath, mod.ModuleName, mod.newModuleName)
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

func renameModule(modulePath string, oldName, newName string) error {
	err := ReplaceInPath(modulePath, []string{"*.go"}, oldName, newName)
	if err != nil {
		return err
	}
	return nil
}

func renameModules(modulePath string, renames map[string]string) error {
	for oldName, newName := range renames {
		err := renameModule(modulePath, oldName, newName)
		if err != nil {
			return err
		}
	}
	return nil
}
