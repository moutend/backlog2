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

var pullRequestCommand = &cobra.Command{
	Use:     "pullrequest",
	Aliases: []string{"pr"},
	RunE: func(c *cobra.Command, args []string) error {
		if err := rootCommand.RunE(c, args); err != nil {
			return err
		}

		return nil
	},
}

var pullRequestListCommand = &cobra.Command{
	Use: "list",
	RunE: func(c *cobra.Command, args []string) error {
		if err := pullRequestCommand.RunE(c, args); err != nil {
			return err
		}
		if len(args) < 2 {
			return nil
		}

		projectKey := args[0]
		repositoryName := args[1]

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

		if err := fetchPullRequests(project.Id, repository.Id); err != nil {
			return err
		}

		pullRequests, err := readPullRequests(project.Id, repository.Id)
		if err != nil {
			return err
		}

		sort.Slice(pullRequests, func(i, j int) bool {
			return pullRequests[i].Number > pullRequests[j].Number
		})

		for _, pullRequest := range pullRequests {
			fmt.Printf("%d. %s (created at %s by %s)\n", pullRequest.Number, pullRequest.Summary, pullRequest.Created.Time().Format("2006-01-02"), pullRequest.CreatedUser.Name)
		}

		return nil
	},
}

func fetchPullRequests(projectId, repositoryId uint64) error {
	q := url.Values{}
	q.Add("projectId", fmt.Sprint(projectId))
	q.Add("repositoryId", fmt.Sprint(repositoryId))

	if time.Now().Sub(lastExecuted(PullRequestsCache, q)) < 30*time.Minute {
		return nil
	}

	pullRequests, err := client.GetPullRequests(fmt.Sprint(projectId), fmt.Sprint(repositoryId), nil)
	if err != nil {
		return err
	}

	base, err := cachePath(PullRequestsCache)
	if err != nil {
		return err
	}

	os.MkdirAll(base, 0755)

	for _, pullRequest := range pullRequests {
		data, err := json.Marshal(pullRequest)
		if err != nil {
			return err
		}

		path := filepath.Join(base, fmt.Sprintf("%d.json", pullRequest.Id))
		if err := ioutil.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}
	if err := setLastExecuted(PullRequestsCache, q); err != nil {
		return err
	}

	return nil
}

func fetchPullRequest(projectId, repositoryId uint64, number string) error {
	pullRequest, err := client.GetPullRequest(fmt.Sprint(projectId), fmt.Sprint(repositoryId), number, nil)
	if err != nil {
		return err
	}

	base, err := cachePath(PullRequestsCache)
	if err != nil {
		return err
	}

	os.MkdirAll(base, 0755)

	data, err := json.Marshal(pullRequest)
	if err != nil {
		return err
	}

	path := filepath.Join(base, fmt.Sprintf("%d.json", pullRequest.Id))
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return err
	}

	return nil
}

func readPullRequests(projectId, repositoryId uint64) (pullRequests []backlog.PullRequest, err error) {
	base, err := cachePath(PullRequestsCache)
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

		var pullRequest backlog.PullRequest

		if err := json.Unmarshal(data, &pullRequest); err != nil {
			return err
		}
		if pullRequest.ProjectId == projectId && pullRequest.RepositoryId == repositoryId {
			pullRequests = append(pullRequests, pullRequest)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return pullRequests, nil
}

func readPullRequest(projectId, repositoryId uint64, number string) (pullRequest backlog.PullRequest, err error) {
	base, err := cachePath(PullRequestsCache)
	if err != nil {
		return pullRequest, err
	}

	err = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		var pr backlog.PullRequest

		if err := json.Unmarshal(data, &pr); err != nil {
			return err
		}
		if pr.ProjectId == projectId && pr.RepositoryId == repositoryId && fmt.Sprint(pr.Number) == number {
			pullRequest = pr
		}

		return nil
	})
	if err != nil {
		return pullRequest, err
	}

	return pullRequest, nil
}

func init() {
	pullRequestCommand.AddCommand(pullRequestListCommand)

	rootCommand.AddCommand(pullRequestCommand)
}
