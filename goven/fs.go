package goven

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func MatchesAny(patterns []string, filename string) (bool, error) {
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

func PrefixedWith(prefixes []string, s string) bool {
	if prefixes != nil {
		for _, prefix := range prefixes {
			if strings.HasPrefix(s, prefix) {
				return true
			}
		}
	}
	return false
}

func CopyFile(sourcePath, destinationPath string) error {
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

func CopyDir(sourcePath, destinationPath string, excludeFilenamePatterns []string, excludePaths []string) error {
	var err = filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf(`failed to copy path "%s" - %s`, path, err.Error())
		}
		var relativePath = strings.Replace(path, sourcePath, "", 1)
		if PrefixedWith(excludePaths, relativePath) {
			return nil
		}
		if info.IsDir() {
			return os.MkdirAll(filepath.Join(destinationPath, relativePath), 0755)
		} else {
			_, filename := filepath.Split(path)
			match, err1 := MatchesAny(excludeFilenamePatterns, filename)
			if err1 != nil {
				return err1
			}
			if !match {
				return CopyFile(filepath.Join(sourcePath, relativePath), filepath.Join(destinationPath, relativePath))
			}
			return nil
		}
	})
	return err
}

func ReplaceInFile(path string, old, new string) error {
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

func ReplaceInPath(source string, filenamePatterns []string, old, new string) error {
	var err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
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
				return ReplaceInFile(path, old, new)
			}
		}
		return nil
	})
	return err
}

func PathExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info != nil
}
