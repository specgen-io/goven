package localize

import (
	"fmt"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"io/ioutil"
	"os"
	"path/filepath"
)

func getLocalModuleName(vendorFolder string, version module.Version) string {
	return fmt.Sprintf(`%s%s%s`, vendorFolder, string(filepath.Separator), version.Path)
}

func renameModule(modulePath string, old, new string) error {
	err := replaceInPath(modulePath, []string{"*.go"}, old, new)
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

func Localize(gomodPath string, newModuleName string, vendorFolder string) error {
	gomod, err := readGomod(gomodPath)
	if err != nil {
		return err
	}
	mainModuleName := gomod.Module.Mod.Path
	moduleRenames := map[string]string{}
	modulePath, _ := filepath.Split(gomodPath)
	os.MkdirAll(filepath.Join(modulePath, vendorFolder), 0755)
	for _, replace := range gomod.Replace {
		sourceModulePath := filepath.Join(modulePath, replace.New.Path)
		localizedPackageName := getLocalModuleName(vendorFolder, replace.Old)
		localizedModulePath := filepath.Join(modulePath, localizedPackageName)
		err := copyDir(sourceModulePath, localizedModulePath, []string{"go.mod", "go.sum"})
		if err != nil {
			return err
		}
		moduleRenames[replace.Old.Path] = fmt.Sprintf(`%s/%s`, mainModuleName, localizedPackageName)
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
