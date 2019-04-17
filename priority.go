package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"time"

	backlog "github.com/moutend/go-backlog"
)

func fetchPriorities() error {
	if time.Now().Sub(lastExecuted(PrioritiesCache, nil)) < 365*24*time.Hour {
		return nil
	}

	priorities, err := client.GetPriorities()
	if err != nil {
		return err
	}

	data, err := json.Marshal(priorities)
	if err != nil {
		return err
	}

	base, err := cachePath(PrioritiesCache)
	if err != nil {
		return err
	}

	path := filepath.Join(base, "priorities.json")
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return err
	}
	if err := setLastExecuted(PrioritiesCache, nil); err != nil {
		return err
	}

	return nil
}

func readPriorities() (priorities []backlog.Priority, err error) {
	base, err := cachePath(PrioritiesCache)
	if err != nil {
		return nil, err
	}

	path := filepath.Join(base, "priorities.json")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &priorities); err != nil {
		return nil, err
	}

	return priorities, nil
}
