package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	backlog "github.com/moutend/go-backlog"
	"github.com/spf13/cobra"
)

var pullRequestCommand = &cobra.Command{
	Use: "pullrequest",
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

		if err := fetchRepositories(project.Id); err != nil {
			return err
		}

		repositories, err := readRepositories(project.Id)
		if err != nil {
			return err
		}

		var repository backlog.Repository

		for _, repository = range repositories {
			if repository.Name == repositoryName {
				break
			}
		}
		if err := fetchPullRequests(fmt.Sprint(project.Id), fmt.Sprint(repository.Id)); err != nil {
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

func fetchPullRequests(projectKeyOrId, repositoryNameOrId string) error {
	prs, err := client.GetPullRequests(projectKeyOrId, repositoryNameOrId, nil)
	if err != nil {
		return err
	}

	base, err := cachePath(PullRequestsCache)
	if err != nil {
		return err
	}

	os.MkdirAll(base, 0755)

	for _, pr := range prs {
		data, err := json.Marshal(pr)
		if err != nil {
			return err
		}

		path := filepath.Join(base, fmt.Sprintf("%d.json", pr.Id))
		if err := ioutil.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

func fetchPullRequest(projectKeyOrId, repositoryNameOrId string, number int) error {
	pr, err := client.GetPullRequest(projectKeyOrId, repositoryNameOrId, number, nil)
	if err != nil {
		return err
	}

	base, err := cachePath(PullRequestsCache)
	if err != nil {
		return err
	}

	os.MkdirAll(base, 0755)

	data, err := json.Marshal(pr)
	if err != nil {
		return err
	}

	path := filepath.Join(base, fmt.Sprintf("%d.json", pr.Id))
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return err
	}

	return nil
}

func readPullRequests(projectId, repositoryId uint64) (prs []backlog.PullRequest, err error) {
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

		var pr backlog.PullRequest

		if err := json.Unmarshal(data, &pr); err != nil {
			return err
		}
		if pr.ProjectId == projectId && pr.RepositoryId == repositoryId {
			prs = append(prs, pr)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return prs, nil
}

func readPullRequest(projectId, repositoryId uint64, number int) (pullRequest backlog.PullRequest, err error) {
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
		if pr.ProjectId == projectId && pr.RepositoryId == repositoryId && pr.Number == number {
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
