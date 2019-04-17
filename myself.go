package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"time"

	backlog "github.com/moutend/go-backlog"
)

func fetchMyself() error {
	if time.Now().Sub(lastExecuted(MyselfCache, nil)) < 365*24*time.Hour {
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

	base, err := cachePath(MyselfCache)
	if err != nil {
		return err
	}

	path := filepath.Join(base, "myself.json")
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return err
	}
	if err := setLastExecuted(MyselfCache, nil); err != nil {
		return err
	}

	return nil
}

func readMyself() (myself backlog.User, err error) {
	base, err := cachePath(MyselfCache)
	if err != nil {
		return myself, err
	}

	path := filepath.Join(base, "myself.json")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return myself, err
	}
	if err := json.Unmarshal(data, &myself); err != nil {
		return myself, err
	}

	return myself, nil
}
