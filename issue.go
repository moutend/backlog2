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

var issueCommand = &cobra.Command{
	Use:     "issue",
	Aliases: []string{"i"},
	RunE: func(c *cobra.Command, args []string) error {
		return nil
	},
}

var (
	issueListMyselfFlag bool
)
var issueListCommand = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	RunE: func(c *cobra.Command, args []string) error {
		var assignee backlog.User

		if issueListMyselfFlag {
			err := fetchMyself()
			if err != nil {
				return err
			}

			assignee, err = readMyself()
			if err != nil {
				return err
			}
		}

		if err := fetchProjects(); err != nil {
			return err
		}

		projects, err := readProjects()
		if err != nil {
			return err
		}

		query := url.Values{}
		query.Add("sort", "updated")
		query.Add("order", "desc")

		if assignee.Id != 0 {
			query.Add("assigneeId[]", fmt.Sprint(assignee.Id))
		}

		for _, project := range projects {
			if err := fetchIssues(project.Id, query); err != nil {
				return err
			}

			issues, err := readIssues(project.Id)
			if err != nil {
				return err
			}

			sort.Slice(issues, func(i, j int) bool {
				return issues[i].Updated.Time().After(issues[j].Updated.Time())
			})

			fmt.Printf("- [%s] %s\n", project.ProjectKey, project.Name)

			for _, issue := range issues {
				if assignee.Id != 0 && assignee.Id != issue.Assignee.Id {
					continue
				}

				fmt.Printf(
					"  - [%s] (%s) %s (updated at %s by %s)\n",
					issue.IssueKey,
					issue.Status.Name,
					issue.Summary,
					issue.Updated.Time().Format("2006-01-02"),
					issue.UpdatedUser.Name,
				)
			}
		}

		return nil
	},
}

var issueShowCommand = &cobra.Command{
	Use:     "show",
	Aliases: []string{"s"},
	RunE: func(c *cobra.Command, args []string) error {
		if len(args) < 1 {
			return nil
		}

		issueKey := args[0]

		if err := fetchIssue(issueKey); err != nil {
			return err
		}

		issue, err := readIssue(issueKey)
		if err != nil {
			return err
		}

		if err := fetchProjectById(issue.ProjectId); err != nil {
			return err
		}

		project, err := readProjectById(issue.ProjectId)
		if err != nil {
			return err
		}

		var parentIssue backlog.Issue

		if issue.ParentIssueId != 0 {
			if err := fetchIssue(fmt.Sprint(issue.ParentIssueId)); err != nil {
				return err
			}

			parentIssue, err = readIssue(fmt.Sprint(issue.ParentIssueId))
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
	Use:     "update",
	Aliases: []string{"u"},
	RunE: func(c *cobra.Command, args []string) error {
		if len(args) < 2 {
			return nil
		}

		issueKey := args[0]
		filePath := args[1]

		query, err := parseIssueMarkdown(issueKey, filePath)
		if err != nil {
			return err
		}

		issue, err := client.UpdateIssue(issueKey, query)
		if err != nil {
			return err
		}

		fmt.Println("updated", issue.IssueKey)

		return nil
	},
}

var issueCreateCommand = &cobra.Command{
	Use:     "create",
	Aliases: []string{"c"},
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

	if time.Now().Sub(lastExecuted(IssuesCache, query)) < 5*time.Minute {
		return nil
	}

	issues, err := client.GetIssues(query)
	if err != nil {
		return err
	}

	base, err := cachePath(IssuesCache)
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
	if err := setLastExecuted(IssuesCache, query); err != nil {
		return err
	}

	return nil
}

func fetchIssue(issueKey string) error {
	issue, err := client.GetIssue(issueKey)
	if err != nil {
		return err
	}

	base, err := cachePath(IssuesCache)
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

func readIssues(projectId uint64) (issues []backlog.Issue, err error) {
	base, err := cachePath(IssuesCache)
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

func readIssue(issueKeyOrId string) (issue backlog.Issue, err error) {
	base, err := cachePath(IssuesCache)
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
		if i.IssueKey == issueKeyOrId || fmt.Sprint(i.Id) == issueKeyOrId {
			issue = i
		}

		return nil
	})

	return issue, nil
}

func init() {
	issueListCommand.Flags().BoolVarP(&issueListMyselfFlag, "myself", "m", false, "pick issues assigned to myself")

	issueCommand.AddCommand(issueListCommand)
	issueCommand.AddCommand(issueShowCommand)
	issueCommand.AddCommand(issueUpdateCommand)
	issueCommand.AddCommand(issueCreateCommand)

	rootCommand.AddCommand(issueCommand)
}
