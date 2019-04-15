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

var pullRequestCommand = &cobra.Command{
	Use: "pullrequest",
	RunE: func(c *cobra.Command, args []string) error {
		if err := rootCommand.RunE(c, args); err != nil {
			return err
		}

		path, err := cachePath(pullRequestsCachePath)
		if err != nil {
			return err
		}

		os.MkdirAll(path, 0755)

		return nil
	},
}

var pullRequestListCommand = &cobra.Command{
	Use: "list",
	RunE: func(c *cobra.Command, args []string) error {
		if err := pullRequestCommand.RunE(c, args); err != nil {
			return err
		}

		if err := fetchProjects(); err != nil {
			return err
		}

		projects, err := readProjects()
		if err != nil {
			return err
		}

		repoM := make(map[uint64][]backlog.Repository)

		for _, project := range projects {
			repos, err := readRepositories(project.Id)
			if err != nil {
				return err
			}

			repoM[project.Id] = repos
		}

		for projectId, repositories := range repoM {
			for _, repository := range repositories {
				if err := fetchPullRequests(projectId, repository.Id); err != nil {
					return err
				}

				prs, err := readPullRequests(projectId, repository.Id)
				if err != nil {
					return err
				}
				for _, pr := range prs {
					fmt.Printf("- %s\n", pr.Summary)
				}

			}
		}

		return nil
	},
}

func fetchPullRequests(projectId, repositoryId uint64) error {
	prs, err := client.GetPullRequests(projectId, repositoryId, nil)
	if err != nil {
		return err
	}

	base, err := cachePath(pullRequestsCachePath)
	if err != nil {
		return err
	}

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

func readPullRequests(projectId, repositoryId uint64) ([]backlog.PullRequest, error) {
	var prs []backlog.PullRequest

	base, err := cachePath(pullRequestsCachePath)
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

func init() {
	pullRequestCommand.AddCommand(pullRequestListCommand)

	rootCommand.AddCommand(pullRequestCommand)
}
