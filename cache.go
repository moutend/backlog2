package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"
)

const (
	cacheDir = ".backlog"

	projectsCachePath = iota
	myselfCachePath
	pullRequestsCachePath
	prioritiesCachePath
	repositoriesCachePath
	statusesCachePath
	issuesCachePath
	issueTypesCachePath
	issueCommentsCachePath
	wikisCachePath
)

func cachePath(t int) (path string, err error) {
	switch t {
	case myselfCachePath:
		path = filepath.Join(cacheDir, "cache", space, "myself.json")
	case projectsCachePath:
		path = filepath.Join(cacheDir, "cache", space, "projects")
	case pullRequestsCachePath:
		path = filepath.Join(cacheDir, "cache", space, "pullrequests")
	case prioritiesCachePath:
		path = filepath.Join(cacheDir, "cache", space, "priorities.json")
	case repositoriesCachePath:
		path = filepath.Join(cacheDir, "cache", space, "repositories")
	case statusesCachePath:
		path = filepath.Join(cacheDir, "cache", space, "statuses.json")
	case issuesCachePath:
		path = filepath.Join(cacheDir, "cache", space, "issues")
	case issueTypesCachePath:
		path = filepath.Join(cacheDir, "cache", space, "issuetypes")
	case issueCommentsCachePath:
		path = filepath.Join(cacheDir, "cache", space, "issue_comments")
	case wikisCachePath:
		path = filepath.Join(cacheDir, "cache", space, "wikis")
	default:
		err = fmt.Errorf("unnown type")
	}

	return path, err
}

func lastExecutedPath(t int) (path string, err error) {
	switch t {
	case myselfCachePath:
		path = filepath.Join(cacheDir, "cache", space, "myself.time")
	case projectsCachePath:
		path = filepath.Join(cacheDir, "cache", space, "project.time")
	case prioritiesCachePath:
		path = filepath.Join(cacheDir, "cache", space, "priorities.time")
	case statusesCachePath:
		path = filepath.Join(cacheDir, "cache", space, "statuses.time")
	case repositoriesCachePath:
		path = filepath.Join(cacheDir, "cache", space, "repository.time")
	case wikisCachePath:
		path = filepath.Join(cacheDir, "cache", space, "wiki.time")
	default:
		err = fmt.Errorf("unknown type")
	}

	return path, err
}

func lastExecuted(t int) time.Time {
	path, err := lastExecutedPath(t)
	if err != nil {
		return time.Time{}
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return time.Time{}
	}

	v, _ := time.Parse(time.RFC3339, string(data))

	return v
}

func setLastExecuted(t int) error {
	path, err := lastExecutedPath(t)
	if err != nil {
		return err
	}

	data := []byte(time.Now().Format(time.RFC3339))
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return err
	}

	return nil
}
