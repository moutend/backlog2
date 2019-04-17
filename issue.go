package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	backlog "github.com/moutend/go-backlog"
	"github.com/spf13/cobra"
)

type ProjectIssues struct {
	Project backlog.Project
	Issues  []backlog.Issue
}

var issueCommand = &cobra.Command{
	Use: "issue",
	RunE: func(c *cobra.Command, args []string) error {
		if err := rootCommand.RunE(c, args); err != nil {
			return err
		}

		return nil
	},
}

var issueListCommand = &cobra.Command{
	Use: "list",
	RunE: func(c *cobra.Command, args []string) error {
		if err := issueCommand.RunE(c, args); err != nil {
			return err
		}

		if err := fetchProjects(); err != nil {
			return err
		}

		projects, err := readProjects()
		if err != nil {
			return err
		}

		pis := make([]ProjectIssues, len(projects))
		query := url.Values{}
		query.Add("sort", "updated")
		query.Add("order", "desc")

		for i, project := range projects {
			if err := fetchIssues(project.Id, query); err != nil {
				return err
			}

			issues, err := readIssues(project.Id)
			if err != nil {
				return err
			}

			sort.Slice(issues, func(i, j int) bool {
				return issues[i].Id > issues[j].Id
			})
			pis[i] = ProjectIssues{
				Project: project,
				Issues:  issues,
			}
		}

		for _, pi := range pis {
			fmt.Printf("- [%s] %s\n", pi.Project.ProjectKey, pi.Project.Name)

			for _, issue := range pi.Issues {
				fmt.Printf(
					"  - [%s] (%s) %s (by %s)\n",
					issue.IssueKey,
					issue.Status.Name,
					issue.Summary,
					issue.CreatedUser.Name,
				)
			}
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
		if err := fetchIssueByIssueKey(args[0]); err != nil {
			return err
		}

		issue, err := readIssueByIssueKey(args[0])
		if err != nil {
			return err
		}

		if err := fetchProjects(); err != nil {
			return err
		}

		projects, err := readProjects()
		if err != nil {
			return err
		}

		var project backlog.Project

		for _, project = range projects {
			if issue.ProjectId == project.Id {
				break
			}
		}

		var parentIssue backlog.Issue

		if issue.ParentIssueId != 0 {
			if err := fetchIssueById(issue.ParentIssueId); err != nil {
				return err
			}

			parentIssue, err = readIssueById(issue.ParentIssueId)
			if err != nil {
				return err
			}
		}

		{
			fmt.Println("---")
			fmt.Println("summary:", issue.Summary)
			fmt.Println("project:", project.ProjectKey)
			if issue.ParentIssueId != 0 {
				fmt.Println("parent:", parentIssue.IssueKey)
			}
			fmt.Println("type:", issue.IssueType.Name)
			fmt.Println("status:", issue.Status.Name)
			fmt.Println("priority:", issue.Priority.Name)
			fmt.Println("assignee:", issue.Assignee.Name)
			fmt.Println("created:", issue.CreatedUser.Name)
			if issue.StartDate.Time().Equal(time.Time{}) {
				fmt.Println("start: ")
			} else {
				fmt.Println("start:", issue.StartDate.Time().Format("2006-01-02"))
			}
			if issue.DueDate.Time().Equal(time.Time{}) {
				fmt.Println("due: ")
			} else {
				fmt.Println("due:", issue.DueDate.Time().Format("2006-01-02"))
			}
			fmt.Println("estimated:", issue.EstimatedHours)
			fmt.Println("actual:", issue.ActualHours)
			fmt.Printf("url: https://%s.backlog.jp/view/%s\n", space, issue.IssueKey)
			fmt.Println("---")
			fmt.Printf("%s", issue.Description)
		}
		return nil
	},
}

var issueUpdateCommand = &cobra.Command{
	Use: "update",
	RunE: func(c *cobra.Command, args []string) error {
		if err := issueCommand.RunE(c, args); err != nil {
			return err
		}
		if len(args) < 2 {
			return nil
		}

		issueKey := args[0]
		filePath := args[1]

		v, err := parseIssueMarkdown(issueKey, filePath)
		if err != nil {
			return err
		}
		fmt.Println(v)
		return nil
	},
}

var issueCreateCommand = &cobra.Command{
	Use: "create",
	RunE: func(c *cobra.Command, args []string) error {
		if err := issueCommand.RunE(c, args); err != nil {
			return err
		}
		if len(args) < 1 {
			return nil
		}

		filePath := args[0]

		v, err := parseIssueMarkdown("", filePath)
		if err != nil {
			return err
		}

		fmt.Println(v)

		return nil
	},
}

func fetchIssues(projectId uint64, query url.Values) error {
	query.Set("count", "100")
	query.Set("projectId[]", fmt.Sprintf("%d", projectId))

	issues, err := client.GetIssues(query)
	if err != nil {
		return err
	}

	base, err := cachePath(issuesCachePath)
	if err != nil {
		return err
	}

	os.MkdirAll(base, 0755)

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

func readIssues(projectId uint64) (issues []backlog.Issue, err error) {
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
		if issue.ProjectId == projectId {
			issues = append(issues, issue)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return issues, nil
}

func fetchIssueByIssueKey(issueKey string) error {
	issue, err := client.GetIssue(issueKey)
	if err != nil {
		return err
	}

	base, err := cachePath(issuesCachePath)
	if err != nil {
		return err
	}

	os.MkdirAll(base, 0755)

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

func fetchIssueById(issueId uint64) error {
	return fetchIssueByIssueKey(fmt.Sprint(issueId))
}

func readIssueByIssueKey(issueKey string) (issue backlog.Issue, err error) {
	base, err := cachePath(issuesCachePath)
	if err != nil {
		return issue, err
	}

	err = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		var i backlog.Issue

		if err := json.Unmarshal(data, &i); err != nil {
			return err
		}
		if i.IssueKey == issueKey {
			issue = i
		}

		return nil
	})

	return issue, nil
}

func readIssueById(issueId uint64) (issue backlog.Issue, err error) {
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

func init() {
	issueCommand.AddCommand(issueListCommand)
	issueCommand.AddCommand(issueShowCommand)
	issueCommand.AddCommand(issueUpdateCommand)
	issueCommand.AddCommand(issueCreateCommand)

	rootCommand.AddCommand(issueCommand)
}
