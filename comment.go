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
	ProjectId  uint64          `json:"projectId"`
	Repository string          `json:"repository"`
	Comment    backlog.Comment `json:"comment"`
}

var commentCommand = &cobra.Command{
	Use: "comment",
	RunE: func(c *cobra.Command, args []string) error {
		if err := rootCommand.RunE(c, args); err != nil {
			return err
		}

		return nil
	},
}

var commentShowCommand = &cobra.Command{
	Use: "show",
	RunE: func(c *cobra.Command, args []string) error {
		if err := commentCommand.RunE(c, args); err != nil {
			return err
		}
		if len(args) == 0 {
			return nil
		}

		var comments []backlog.Comment

		switch len(args) {
		case 1: // issue
			if err := fetchIssueByIssueKey(args[0]); err != nil {
				return err
			}

			issue, err := readIssueByIssueKey(args[0])
			if err != nil {
				return err
			}

			if err := fetchCommentsByIssueId(issue.Id); err != nil {
				return err
			}

			comments, err = readCommentsByIssueId(issue.Id)
			if err != nil {
				return err
			}
			break
		case 2: // pull-request
			break
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

func fetchCommentsByIssueId(issueId uint64) error {
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

func readCommentsByIssueId(issueId uint64) (comments []backlog.Comment, err error) {
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

		comments = append(comments, ic.Comment)

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
