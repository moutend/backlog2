package main

import (
	"fmt"
	"path/filepath"
)

const (
	cacheDir = ".backlog"

	projectsCachePath = iota
	issuesCachePath
)

func cachePath(t int) (path string, err error) {
	switch t {
	case projectsCachePath:
		path = filepath.Join(cacheDir, "cache", space, "projects")
	case issuesCachePath:
		path = filepath.Join(cacheDir, "cache", space, "issues")
	default:
		err = fmt.Errorf("path is not defined")
	}

	return path, err
}
