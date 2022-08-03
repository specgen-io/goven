package goven

import (
	"fmt"
	"os"
	"path/filepath"
)

func vendorModule(sourceModulePath, targetPath, vendoredModulesFolder, oldModuleName string) (string, error) {
	vendoredPackageName := filepath.Join(vendoredModulesFolder, oldModuleName)
	vendoredModulePath := filepath.Join(targetPath, vendoredPackageName)

	err := CopyDir(sourceModulePath, vendoredModulePath, []string{"go.mod", "go.sum"})
	if err != nil {
		return "", err
	}
	return vendoredPackageName, nil
}

func Vendor(gomodPath, outputPath, newModuleName, vendoredModulesFolder string, vendorRequired bool) error {
	gomodPath, err := filepath.Abs(gomodPath)
	if err != nil {
		return fmt.Errorf(`can't get absolute path for "%s" - %s`, gomodPath, err.Error())
	}
	modulePath, modFilename := filepath.Split(gomodPath)
	if vendorRequired {
		modVendorPath := filepath.Join(modulePath, "vendor")
		if _, err := os.Stat(modVendorPath); os.IsNotExist(err) {
			return fmt.Errorf(`can't find vendor folder: "%s", run "go mod vendor" in the module first'`, modVendorPath)
		}
	}
	if outputPath != "" {
		err := CopyDir(modulePath, outputPath, nil)
		if err != nil {
			return err
		}
	} else {
		outputPath = modulePath
	}

	mod, err := OpenModfile(outputPath, modFilename)
	if err != nil {
		return err
	}

	for _, replace := range mod.File.Replace {
		sourceModulePath := filepath.Join(modulePath, replace.New.Path)
		oldModuleName := replace.Old.Path
		vendoredModuleName, err := vendorModule(sourceModulePath, outputPath, vendoredModulesFolder, oldModuleName)
		if err != nil {
			return err
		}
		mod.RenameSubmodule(oldModuleName, fmt.Sprintf(`%s/%s`, mod.ModuleName, vendoredModuleName))
		mod.DropRequire(oldModuleName)
		mod.DropReplace(oldModuleName, replace.Old.Version)
	}

	if vendorRequired {
		modVendorPath := filepath.Join(outputPath, "vendor")
		for _, require := range mod.File.Require {
			if !mod.IsRequireDropped(require.Mod.Path) {
				sourceModulePath := filepath.Join(modVendorPath, require.Mod.Path)
				oldModuleName := require.Mod.Path
				vendoredModuleName, err := vendorModule(sourceModulePath, outputPath, vendoredModulesFolder, oldModuleName)
				if err != nil {
					return err
				}
				mod.RenameSubmodule(oldModuleName, fmt.Sprintf(`%s/%s`, mod.ModuleName, vendoredModuleName))
				mod.DropRequire(oldModuleName)
			}
		}
		err = os.RemoveAll(modVendorPath)
		if err != nil {
			return err
		}
	}

	mod.Rename(newModuleName)

	err = mod.Save("go.mod")
	if err != nil {
		return err
	}
	return nil
}
