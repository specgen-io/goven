package localize

import (
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
			return os.MkdirAll(filepath.Join(destinationPath, relativePath), 0755)
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
	var err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
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
