package git

import (
	"fmt"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
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

	_, err = w.Remove(path)
	if err != nil {
		return fmt.Errorf(`removing path "%s" from working copy failed: %s`, path, err.Error())
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
		_, err = r.CreateTag(tag, commit, &git.CreateTagOptions{Message: commitMessage, Tagger: signature})
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
