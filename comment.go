package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	backlog "github.com/moutend/go-backlog"
	"github.com/spf13/cobra"
)

type IssueComment struct {
	IssueId uint64          `json:"issueId"`
	Comment backlog.Comment `json:"comment"`
}

type PullRequestComment struct {
	ProjectId    uint64          `json:"projectId"`
	RepositoryId uint64          `json:"repositoryId"`
	Number       string          `json:"number"`
	Comment      backlog.Comment `json:"comment"`
}

var commentCommand = &cobra.Command{
	Use:     "comment",
	Aliases: []string{"c"},
	RunE: func(c *cobra.Command, args []string) error {
		return nil
	},
}

var commentShowCommand = &cobra.Command{
	Use: "show",
	RunE: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			return nil
		}

		var comments []backlog.Comment

		switch len(args) {
		case 1: // issue
			issueKey := args[0]

			if err := fetchIssue(issueKey); err != nil {
				return err
			}

			issue, err := readIssue(issueKey)
			if err != nil {
				return err
			}

			if err := fetchIssueComments(issue.Id); err != nil {
				return err
			}

			comments, err = readIssueComments(issue.Id)
			if err != nil {
				return err
			}
		case 3: // pull-request
			projectKey := args[0]
			repositoryName := args[1]
			number := args[2]

			if err := fetchProjectByProjectKey(projectKey); err != nil {
				return err
			}

			project, err := readProjectByProjectKey(projectKey)
			if err != nil {
				return err
			}

			if err := fetchRepository(project.Id, repositoryName); err != nil {
				return err
			}

			repository, err := readRepository(project.Id, repositoryName)
			if err != nil {
				return err
			}

			if err := fetchPullRequestComments(project.Id, repository.Id, number); err != nil {
				return err
			}

			comments, err = readPullRequestComments(project.Id, repository.Id, number)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("specify issue or pull-request")
		}
		for i, _ := range comments {
			comment := comments[len(comments)-i-1]
			if len(comment.ChangeLog) > 0 {
				fmt.Println(comment.CreatedUser.Name, "が課題の内容を変更しました。")
				fmt.Println(comment.Content)
				for _, change := range comment.ChangeLog {
					fmt.Println("  ", change.Field, change.OriginalValue, "->", change.NewValue)
				}
			} else {
				fmt.Println(comment.CreatedUser.Name, comment.Content)
			}
		}
		return nil
	},
}

func fetchIssueComments(issueId uint64) error {
	comments, err := client.GetIssueComments(issueId, nil)
	if err != nil {
		return err
	}

	base, err := cachePath(IssueCommentsCache)
	if err != nil {
		return err
	}

	os.MkdirAll(base, 0755)

	for _, comment := range comments {
		c := IssueComment{
			IssueId: issueId,
			Comment: comment,
		}
		data, err := json.Marshal(c)
		if err != nil {
			return err
		}

		path := filepath.Join(base, fmt.Sprintf("%d.json", comment.Id))
		if err := ioutil.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

func fetchPullRequestComments(projectId, repositoryId uint64, number string) error {
	comments, err := client.GetPullRequestComments(fmt.Sprint(projectId), fmt.Sprint(repositoryId), number, nil)
	if err != nil {
		return err
	}

	base, err := cachePath(PullRequestCommentsCache)
	if err != nil {
		return err
	}

	os.MkdirAll(base, 0755)

	for _, comment := range comments {
		c := PullRequestComment{
			ProjectId:    projectId,
			RepositoryId: repositoryId,
			Number:       number,
			Comment:      comment,
		}
		data, err := json.Marshal(c)
		if err != nil {
			return err
		}

		path := filepath.Join(base, fmt.Sprintf("%d.json", comment.Id))
		if err := ioutil.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

func readIssueComments(issueId uint64) (comments []backlog.Comment, err error) {
	base, err := cachePath(IssueCommentsCache)
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

		var ic IssueComment

		if err := json.Unmarshal(data, &ic); err != nil {
			return err
		}
		if ic.IssueId == issueId {
			comments = append(comments, ic.Comment)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func readPullRequestComments(projectId, repositoryId uint64, number string) (comments []backlog.Comment, err error) {
	base, err := cachePath(PullRequestCommentsCache)
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

		var prc PullRequestComment

		if err := json.Unmarshal(data, &prc); err != nil {
			return err
		}
		if prc.ProjectId == projectId && prc.RepositoryId == repositoryId && prc.Number == number {
			comments = append(comments, prc.Comment)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func init() {
	commentCommand.AddCommand(commentShowCommand)

	rootCommand.AddCommand(commentCommand)
}
