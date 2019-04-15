package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	backlog "github.com/moutend/go-backlog"
	"github.com/spf13/cobra"
)

var repositoryCommand = &cobra.Command{
	Use: "repository",
	RunE: func(c *cobra.Command, args []string) error {
		if err := rootCommand.RunE(c, args); err != nil {
			return err
		}

		path, err := cachePath(repositoriesCachePath)
		if err != nil {
			return err
		}

		os.MkdirAll(path, 0755)

		return nil
	},
}

var repositoryListCommand = &cobra.Command{
	Use: "list",
	RunE: func(c *cobra.Command, args []string) error {
		if err := repositoryCommand.RunE(c, args); err != nil {
			return err
		}

		if err := fetchProjects(); err != nil {
			return err
		}

		projects, err := readProjects()
		if err != nil {
			return err
		}

		for _, project := range projects {
			if err := fetchRepositories(project.Id); err != nil {
				return err
			}

			repositories, err := readRepositories(project.Id)
			if err != nil {
				return err
			}

			fmt.Printf("- [%s] %s\n", project.ProjectKey, project.Name)

			for _, repository := range repositories {
				fmt.Printf("  - %v\n", repository.Name)
			}
		}

		return nil
	},
}

func fetchRepositories(projectId uint64) error {
	if time.Now().Sub(lastExecuted(repositoriesCachePath)) < 24*time.Hour {
		return nil
	}

	repositories, err := client.GetRepositories(projectId, nil)
	if err != nil {
		return err
	}

	base, err := cachePath(repositoriesCachePath)
	if err != nil {
		return err
	}

	for _, repository := range repositories {
		data, err := json.Marshal(repository)
		if err != nil {
			return err
		}

		path := filepath.Join(base, fmt.Sprintf("%d.json", repository.Id))
		if err := ioutil.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}

	if err := setLastExecuted(repositoriesCachePath); err != nil {
		return err
	}

	return nil
}

func readRepositories(projectId uint64) ([]backlog.Repository, error) {
	var repositories []backlog.Repository

	base, err := cachePath(repositoriesCachePath)
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

		var repository backlog.Repository

		if err := json.Unmarshal(data, &repository); err != nil {
			return err
		}
		if repository.ProjectId != projectId {
			return nil
		}

		repositories = append(repositories, repository)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return repositories, nil
}

func init() {
	repositoryCommand.AddCommand(repositoryListCommand)

	rootCommand.AddCommand(repositoryCommand)
}
