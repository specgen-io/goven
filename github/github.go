package github

import (
	"fmt"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Credentials struct {
	GithubName, GithubEmail, GuthubUser, GithubToken string
}

func PutFiles(sourcePath string, repoUrl, path string, tag string, credentials Credentials) error {
	filesystem := memfs.New()
	r, err := git.Clone(memory.NewStorage(), filesystem, &git.CloneOptions{URL: repoUrl})
	if err != nil {
		return fmt.Errorf(`clone repo "%s" failed: %s`, repoUrl, err.Error())
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf(`getting work tree for repo "%s" failed: %s`, repoUrl, err.Error())
	}

	info, _ := filesystem.Lstat(path)
	if info != nil {
		err = removeDir(filesystem, path)
		if err != nil {
			return err
		}
	}

	err = copyDir(sourcePath, filesystem, path)
	if err != nil {
		return err
	}

	err = w.AddWithOptions(&git.AddOptions{All: true})
	if err != nil {
		return fmt.Errorf(`git add failed: %s`, err.Error())
	}

	commitMessage := "Release changes"
	if tag != "" {
		commitMessage = fmt.Sprintf(`Release %s`, tag)
	}

	signature := &object.Signature{Name: credentials.GithubName, Email: credentials.GithubEmail, When: time.Now()}
	commit, err := w.Commit(commitMessage, &git.CommitOptions{Author: signature})
	if err != nil {
		return fmt.Errorf(`git commit failed: %s`, err.Error())
	}

	auth := &http.BasicAuth{Username: credentials.GuthubUser, Password: credentials.GithubToken}

	err = r.Push(&git.PushOptions{Auth: auth})
	if err != nil {
		return fmt.Errorf(`git push failed: %s`, err.Error())
	}

	if tag != "" {
		_, err = r.CreateTag(tag, commit, &git.CreateTagOptions{Message: commitMessage})
		if err != nil {
			return fmt.Errorf(`create git tag failed: %s`, err.Error())
		}
		err = r.Push(&git.PushOptions{Auth: auth, RefSpecs: []config.RefSpec{"refs/tags/*:refs/tags/*"}})
		if err != nil {
			return fmt.Errorf(`git push tag failed: %s`, err.Error())
		}
	}

	return nil
}

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
