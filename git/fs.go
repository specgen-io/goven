package git

import (
	"fmt"
	"github.com/go-git/go-billy/v5"
	"io"
	"os"
	"path/filepath"
)

func removeDir(filesystem billy.Filesystem, path string) error {
	items, err := filesystem.ReadDir(path)
	if err != nil {
		return fmt.Errorf(`can't list directory "%s": %s`, path, err.Error())
	}
	for _, item := range items {
		itempath := filepath.Join(path, item.Name())
		if item.IsDir() {
			err = removeDir(filesystem, itempath)
			if err != nil {
				return err
			}
		} else {
			err = filesystem.Remove(itempath)
			if err != nil {
				return fmt.Errorf(`can't delete file "%s": %s`, itempath, err.Error())
			}
		}
	}
	err = filesystem.Remove(path)
	if err != nil {
		return fmt.Errorf(`can't delete directory "%s": %s`, path, err.Error())
	}
	return nil
}

func copyDir(sourcePath string, filesystem billy.Filesystem, targetPath string) error {
	var err = filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf(`failed to copy path "%s" - %s`, sourcePath, err.Error())
		}
		relativePath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return fmt.Errorf(`failed to get relative path "%s" - %s`, path, err.Error())
		}
		itemTargetPath := filepath.Join(targetPath, relativePath)
		if info.IsDir() {
			err1 := filesystem.MkdirAll(itemTargetPath, os.ModeDir)
			if err1 != nil {
				return fmt.Errorf(`failed create directory "%s": %s`, itemTargetPath, err1.Error())
				return err1
			}
		} else {
			err1 := copyFile(path, filesystem, itemTargetPath)
			if err1 != nil {
				return fmt.Errorf(`failed to copy file "%s" to "%s": %s`, path, itemTargetPath, err1.Error())
			}
		}
		return nil
	})
	return err
}

func copyFile(sourcePath string, filesystem billy.Filesystem, targetPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf(`failed to open file "%s": %s`, sourcePath, err.Error())
	}
	defer source.Close()

	target, err := filesystem.Create(targetPath)
	if err != nil {
		return fmt.Errorf(`failed to create file "%s": %s`, targetPath, err.Error())
	}
	defer target.Close()

	_, err = io.Copy(target, source)
	if err != nil {
		return fmt.Errorf(`failed to copy data from "%s" to "%s": %s`, sourcePath, targetPath, err.Error())
	}
	return nil
}
