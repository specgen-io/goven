package localize

import (
	"golang.org/x/mod/modfile"
	"io/ioutil"
)

func gomodRead(gomodPath string) (*modfile.File, error) {
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

func gomodWrite(file *modfile.File, path string) error {
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

func gomodApplyChanges(gomod *modfile.File, newModuleName string, moduleRequireDrops map[string]any, moduleReplaceDrops map[string]string) error {
	for moduleName, version := range moduleReplaceDrops {
		err := gomod.DropReplace(moduleName, version)
		if err != nil {
			return err
		}
	}
	for require := range moduleRequireDrops {
		err := gomod.DropRequire(require)
		if err != nil {
			return err
		}
	}
	if newModuleName != "" {
		err := gomod.AddModuleStmt(newModuleName)
		if err != nil {
			return err
		}
	}
	return nil
}
