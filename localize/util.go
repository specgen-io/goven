package localize

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func matchesAny(patterns []string, filename string) (bool, error) {
	if patterns != nil {
		for _, pattern := range patterns {
			match, err := filepath.Match(pattern, filename)
			if err != nil {
				return false, fmt.Errorf(`failed match filename "%s" with patterns "%s" - %s`, filename, strings.Join(patterns, "|"), err.Error())
			}
			if match {
				return true, nil
			}
		}
	}
	return false, nil
}

func copyFile(sourcePath, destinationPath string) error {
	data, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf(`failed to read file "%s" - %s`, sourcePath, err.Error())
	}
	err = ioutil.WriteFile(destinationPath, data, 0777)
	if err != nil {
		return fmt.Errorf(`failed to write file "%s" - %s`, destinationPath, err.Error())
	}
	return nil
}

func copyDir(sourcePath, destinationPath string, excludePatterns []string) error {
	var err = filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf(`failed to copy path "%s" - %s`, path, err.Error())
		}
		var relativePath = strings.Replace(path, sourcePath, "", 1)
		if info.IsDir() {
			return os.MkdirAll(filepath.Join(destinationPath, relativePath), 0755)
		} else {
			_, filename := filepath.Split(path)
			match, err1 := matchesAny(excludePatterns, filename)
			if err1 != nil {
				return err1
			}
			if !match {
				return copyFile(filepath.Join(sourcePath, relativePath), filepath.Join(destinationPath, relativePath))
			}
			return nil
		}
	})
	return err
}

func replaceInFile(path string, old, new string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf(`failed to read file "%s" - %s`, path, err.Error())
	}
	result := strings.Replace(string(data), old, new, -1)
	err = ioutil.WriteFile(path, []byte(result), 0644)
	if err != nil {
		return fmt.Errorf(`failed to write file "%s" - %s`, path, err.Error())
	}
	return nil
}

func replaceInPath(source string, patterns []string, old, new string) error {
	var err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf(`failed to replace in path "%s" - %s`, path, err.Error())
		}
		if !info.IsDir() {
			_, filename := filepath.Split(path)
			match, err1 := matchesAny(patterns, filename)
			if err1 != nil {
				return err1
			}
			if match {
				return replaceInFile(path, old, new)
			}
		}
		return nil
	})
	return err
}
