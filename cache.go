package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
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

func lastExecutedPath(t int, query url.Values) (path string, err error) {
	switch t {
	case myselfCachePath:
		path = filepath.Join(cacheDir, "cache", space, "myself")
	case projectsCachePath:
		path = filepath.Join(cacheDir, "cache", space, "projects")
	case prioritiesCachePath:
		path = filepath.Join(cacheDir, "cache", space, "priorities")
	case statusesCachePath:
		path = filepath.Join(cacheDir, "cache", space, "statuses")
	case repositoriesCachePath:
		path = filepath.Join(cacheDir, "cache", space, "repositories")
	case wikisCachePath:
		path = filepath.Join(cacheDir, "cache", space, "wikis")
	default:
		err = fmt.Errorf("unknown type")
	}
	if hash := hashQuery(query); hash == "" {
		path = fmt.Sprintf("%s.time", path)
	} else {
		path = fmt.Sprintf("%s.%s.time", path, hash)
	}

	return path, err
}

func lastExecuted(t int, query url.Values) time.Time {
	path, err := lastExecutedPath(t, query)
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

func setLastExecuted(t int, query url.Values) error {
	path, err := lastExecutedPath(t, query)
	if err != nil {
		return err
	}

	data := []byte(time.Now().Format(time.RFC3339))
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return err
	}

	return nil
}

func hashQuery(query url.Values) string {
	if query == nil {
		return ""
	}

	ss := []string{}

	for k, v := range query {
		ss = append(ss, fmt.Sprintf("%s:%s", k, v))
	}

	sort.Strings(ss)

	s := strings.Join(ss, ",")

	b := sha256.Sum256([]byte(s))

	return hex.EncodeToString(b[:])
}
