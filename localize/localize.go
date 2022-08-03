package localize

import (
	"fmt"
	"os"
	"path/filepath"
)

func getLocalModuleName(location string, moduleName string) string {
	return filepath.Join(location, moduleName)
}

func renameModule(modulePath string, old, new string) error {
	err := replaceInPath(modulePath, []string{"*.go"}, old, new)
	if err != nil {
		return err
	}
	return nil
}

func renameModules(modulePath string, oldModuleName, newModuleName string, renames map[string]string) error {
	for oldName, newName := range renames {
		err := renameModule(modulePath, oldName, newName)
		if err != nil {
			return err
		}
	}
	if newModuleName != "" {
		err := renameModule(modulePath, oldModuleName, newModuleName)
		if err != nil {
			return err
		}
	}
	return nil
}

func Localize(gomodPath, outputPath, newModuleName, localizedModulesPath string, localizeVendor bool) error {
	gomodPath, err := filepath.Abs(gomodPath)
	if err != nil {
		return fmt.Errorf(`can't get absolute path for "%s" - %s`, gomodPath, err.Error())
	}
	modulePath, gomodFilename := filepath.Split(gomodPath)
	if localizeVendor {
		vendorPath := filepath.Join(modulePath, "vendor")
		if _, err := os.Stat(vendorPath); os.IsNotExist(err) {
			return fmt.Errorf(`can't find vendor folder: "%s", run "go mod vendor" in the module first'`, vendorPath)
		}
	}
	if outputPath != "" {
		err := copyDir(modulePath, outputPath, nil)
		if err != nil {
			return err
		}
	} else {
		outputPath = modulePath
	}
	gomodOutputPath := filepath.Join(outputPath, gomodFilename)

	gomod, err := gomodRead(gomodOutputPath)
	if err != nil {
		return err
	}
	mainModuleName := gomod.Module.Mod.Path
	moduleRenames := map[string]string{}
	moduleRequireDrops := map[string]any{}
	moduleReplaceDrops := map[string]string{}

	for _, replace := range gomod.Replace {
		sourceModulePath := filepath.Join(modulePath, replace.New.Path)
		localizedPackageName := getLocalModuleName(localizedModulesPath, replace.Old.Path)
		localizedModulePath := filepath.Join(outputPath, localizedPackageName)

		err := copyDir(sourceModulePath, localizedModulePath, []string{"go.mod", "go.sum"})
		if err != nil {
			return err
		}
		moduleRenames[replace.Old.Path] = fmt.Sprintf(`%s/%s`, mainModuleName, localizedPackageName)
		moduleRequireDrops[replace.Old.Path] = true
		moduleReplaceDrops[replace.Old.Path] = replace.Old.Version
	}

	if localizeVendor {
		vendorPath := filepath.Join(outputPath, "vendor")
		for _, require := range gomod.Require {
			if _, found := moduleRequireDrops[require.Mod.Path]; !found {
				sourceModulePath := filepath.Join(vendorPath, require.Mod.Path)
				localizedPackageName := getLocalModuleName(localizedModulesPath, require.Mod.Path)
				localizedModulePath := filepath.Join(outputPath, localizedPackageName)

				err := copyDir(sourceModulePath, localizedModulePath, []string{"go.mod", "go.sum"})
				if err != nil {
					return err
				}
				moduleRenames[require.Mod.Path] = fmt.Sprintf(`%s/%s`, mainModuleName, localizedPackageName)
				moduleRequireDrops[require.Mod.Path] = true
			}
		}
		os.RemoveAll(vendorPath)
	}

	err = renameModules(outputPath, mainModuleName, newModuleName, moduleRenames)
	if err != nil {
		return err
	}
	err = gomodApplyChanges(gomod, newModuleName, moduleRequireDrops, moduleReplaceDrops)
	if err != nil {
		return err
	}
	err = gomodWrite(gomod, gomodOutputPath)
	if err != nil {
		return err
	}
	return nil
}
