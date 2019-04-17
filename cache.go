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

const CacheDir = ".backlog"

func cachePath(ct cacheType) (path string, err error) {
	if ct.String() == "" {
		return path, fmt.Errorf("unknown cache type")
	}

	path = filepath.Join(CacheDir, "cache", space, ct.String())

	return path, err
}

func lastExecutedPath(ct cacheType, query url.Values) (path string, err error) {
	if ct.String() == "" {
		return path, fmt.Errorf("unknown cache type")
	}

	path = filepath.Join(CacheDir, "cache", space, ct.String())

	if hash := hashQuery(query); hash == "" {
		path = fmt.Sprintf("%s.time", path)
	} else {
		path = fmt.Sprintf("%s.%s.time", path, hash)
	}

	return path, err
}

func lastExecuted(t cacheType, query url.Values) time.Time {
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

func setLastExecuted(t cacheType, query url.Values) error {
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
