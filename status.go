package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	backlog "github.com/moutend/go-backlog"
)

func fetchStatuses() error {
	if time.Now().Sub(lastExecuted(StatusesCache, nil)) < 365*24*time.Hour {
		return nil
	}
	statuses, err := client.GetStatuses()
	if err != nil {
		return err
	}

	data, err := json.Marshal(statuses)
	if err != nil {
		return err
	}

	base, err := cachePath(StatusesCache)
	if err != nil {
		return err
	}

	os.MkdirAll(base, 0755)

	path := filepath.Join(base, "statuses.json")
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return err
	}
	if err := setLastExecuted(StatusesCache, nil); err != nil {
		return err
	}

	return nil
}

func readStatuses() (statuses []backlog.Status, err error) {
	base, err := cachePath(StatusesCache)
	if err != nil {
		return nil, err
	}

	path := filepath.Join(base, "statuses.json")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &statuses); err != nil {
		return nil, err
	}

	return statuses, nil
}
