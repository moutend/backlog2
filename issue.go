package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
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

var issueShowCommand = &cobra.Command{
	Use: "show",
	RunE: func(c *cobra.Command, args []string) error {
		if err := issueCommand.RunE(c, args); err != nil {
			return err
		}
		if len(args) < 1 {
			return nil
		}
		if err := fetchIssue(args[0]); err != nil {
			return err
		}

		issue, err := readIssue(args[0])
		if err != nil {
			return err
		}

		{
			fmt.Println("---")
			fmt.Println("summary:", issue.Summary)
			fmt.Println("parentissueid:", issue.ParentIssueId)
			fmt.Println("issuetype:", issue.IssueType.Name)
			fmt.Println("status:", issue.Status.Name)
			fmt.Println("priority:", issue.Priority.Name)
			fmt.Println("assignee:", issue.Assignee.Name)
			fmt.Println("created:", issue.CreatedUser.Name)
			fmt.Println("start:", issue.StartDate.Time().Format("2006-01-02"))
			fmt.Println("due:", issue.DueDate.Time().Format("2006-01-02"))
			fmt.Println("estimated:", issue.EstimatedHours)
			fmt.Println("actual:", issue.ActualHours)
			fmt.Printf("url: https://%s.backlog.jp/view/%s\n", space, issue.IssueKey)
			fmt.Println("---")
			fmt.Printf("%s", issue.Description)
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
		query.Add("count", "100")

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

func fetchIssue(key string) error {
	issue, err := client.GetIssue(key)
	if err != nil {
		return err
	}

	base, err := cachePath(issuesCachePath)
	if err != nil {
		return err
	}

	data, err := json.Marshal(issue)
	if err != nil {
		return err
	}

	path := filepath.Join(base, fmt.Sprintf("%d.json", issue.Id))
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return err
	}

	return nil
}

func readIssue(key string) (issue backlog.Issue, err error) {
	issueId, err := strconv.Atoi(key)
	if err != nil {
		return readIssueByKey(key)
	}

	base, err := cachePath(issuesCachePath)
	if err != nil {
		return issue, err
	}

	path := filepath.Join(base, fmt.Sprintf("%d.json", issueId))
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return issue, err
	}

	if err := json.Unmarshal(data, &issue); err != nil {
		return issue, err
	}

	return issue, nil
}

func readIssueByKey(key string) (issue backlog.Issue, err error) {
	issues, err := readIssues()
	if err != nil {
		return issue, err
	}
	for _, issue := range issues {
		if issue.IssueKey == key {
			return issue, nil
		}
	}

	return issue, fmt.Errorf("%s not found", key)
}

func init() {
	issueCommand.AddCommand(issueListCommand)
	issueCommand.AddCommand(issueShowCommand)

	rootCommand.AddCommand(issueCommand)
}
