package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	backlog "github.com/moutend/go-backlog"
)

func fetchIssueTypes(projectId uint64) error {
	issueTypes, err := client.GetIssueTypes(projectId)
	if err != nil {
		return err
	}

	base, err := cachePath(IssueTypesCache)
	if err != nil {
		return err
	}

	os.MkdirAll(base, 0755)

	for _, issueType := range issueTypes {
		data, err := json.Marshal(issueType)
		if err != nil {
			return err
		}

		path := filepath.Join(base, fmt.Sprintf("%d.json", issueType.Id))
		if err := ioutil.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

func readIssueTypes(projectId uint64) (issueTypes []backlog.IssueType, err error) {
	base, err := cachePath(IssueTypesCache)
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		var issueType backlog.IssueType

		if err := json.Unmarshal(data, &issueType); err != nil {
			return err
		}
		if issueType.ProjectId == projectId {
			issueTypes = append(issueTypes, issueType)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return issueTypes, nil
}
