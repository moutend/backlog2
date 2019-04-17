package main

import (
	"encoding/json"
	"io/ioutil"
	"time"

	backlog "github.com/moutend/go-backlog"
)

func fetchStatuses() error {
	if time.Now().Sub(lastExecuted(statusesCachePath, nil)) < 365*24*time.Hour {
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

	path, err := cachePath(statusesCachePath)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return err
	}
	if err := setLastExecuted(statusesCachePath, nil); err != nil {
		return err
	}

	return nil
}

func readStatuses() (statuses []backlog.Status, err error) {
	path, err := cachePath(statusesCachePath)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &statuses); err != nil {
		return nil, err
	}

	return statuses, nil
}
