package main

import (
	"encoding/json"
	"io/ioutil"
	"time"

	backlog "github.com/moutend/go-backlog"
)

func fetchMyself() error {
	if time.Now().Sub(lastExecuted(myselfCachePath)) < 365*24*time.Hour {
		return nil
	}

	myself, err := client.GetMyself()
	if err != nil {
		return err
	}

	data, err := json.Marshal(myself)
	if err != nil {
		return err
	}

	path, err := cachePath(myselfCachePath)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return err
	}
	if err := setLastExecuted(myselfCachePath); err != nil {
		return err
	}

	return nil
}

func readMyself() (myself backlog.User, err error) {
	path, err := cachePath(myselfCachePath)
	if err != nil {
		return myself, err
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return myself, err
	}
	if err := json.Unmarshal(data, &myself); err != nil {
		return myself, err
	}

	return myself, nil
}
