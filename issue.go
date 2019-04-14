package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	backlog "github.com/moutend/go-backlog"
	"github.com/spf13/cobra"
)

var issueCommand = &cobra.Command{
	Use: "issue",
	RunE: func(c *cobra.Command, args []string) error {
		if err := rootCommand.RunE(c, args); err != nil {
			return err
		}

		path, err := cachePath(issuesCachePath)
		if err != nil {
			return err
		}

		os.MkdirAll(path, 0755)

		return nil
	},
}

var issueListCommand = &cobra.Command{
	Use: "list",
	RunE: func(c *cobra.Command, args []string) error {
		if err := issueCommand.RunE(c, args); err != nil {
			return err
		}
		if err := fetchIssues(); err != nil {
			return err
		}

		issues, err := readIssues()
		if err != nil {
			return err
		}

		for _, issue := range issues {
			fmt.Printf("  - [%v %v] %v by @%v (id:%v) %v\n", issue.Status.Name, issue.IssueType.Name, issue.Summary, issue.CreatedUser.Name, issue.IssueKey, issue.StartDate)
		}

		return nil
	},
}

func fetchIssues() error {
	if err := fetchProjects(); err != nil {
		return err
	}

	projects, err := readProjects()
	if err != nil {
		return err
	}

	var issues []backlog.Issue

	for _, project := range projects {
		query := url.Values{}
		query.Add("projectId[]", fmt.Sprintf("%d", project.Id))

		i, err := client.GetIssues(query)
		if err != nil {
			return err
		}

		issues = append(issues, i...)
	}

	base, err := cachePath(issuesCachePath)
	if err != nil {
		return err
	}

	for _, issue := range issues {
		data, err := json.Marshal(issue)
		if err != nil {
			return err
		}

		path := filepath.Join(base, fmt.Sprintf("%d.json", issue.Id))
		if err := ioutil.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

func readIssues() ([]backlog.Issue, error) {
	var issues []backlog.Issue

	base, err := cachePath(issuesCachePath)
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

		var issue backlog.Issue

		if err := json.Unmarshal(data, &issue); err != nil {
			return err
		}

		issues = append(issues, issue)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return issues, nil
}

func init() {
	issueCommand.AddCommand(issueListCommand)
	rootCommand.AddCommand(issueCommand)
}
