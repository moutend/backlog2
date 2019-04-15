package main

import (
	"fmt"
	"path/filepath"
)

const (
	cacheDir = ".backlog"

	projectsCachePath = iota
	repositoriesCachePath
	issuesCachePath
	issueCommentsCachePath
)

func cachePath(t int) (path string, err error) {
	switch t {
	case projectsCachePath:
		path = filepath.Join(cacheDir, "cache", space, "projects")
	case repositoriesCachePath:
		path = filepath.Join(cacheDir, "cache", space, "repositories")
	case issuesCachePath:
		path = filepath.Join(cacheDir, "cache", space, "issues")
	case issueCommentsCachePath:
		path = filepath.Join(cacheDir, "cache", space, "issue_comments")
	default:
		err = fmt.Errorf("unnown type")
	}

	return path, err
}
