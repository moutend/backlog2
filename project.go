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

var projectCommand = &cobra.Command{
	Use: "project",
	RunE: func(c *cobra.Command, args []string) error {
		if err := rootCommand.RunE(c, args); err != nil {
			return err
		}

		path, err := cachePath(projectsCachePath)
		if err != nil {
			return err
		}

		os.MkdirAll(path, 0755)

		return nil
	},
}

var projectListCommand = &cobra.Command{
	Use: "list",
	RunE: func(c *cobra.Command, args []string) error {
		if err := projectCommand.RunE(c, args); err != nil {
			return err
		}
		if err := fetchProjects(); err != nil {
			return err
		}

		projects, err := readProjects()
		if err != nil {
			return err
		}

		for i, project := range projects {
			fmt.Printf("%d. [%s] %s\n", i+1, project.ProjectKey, project.Name)
		}

		return nil
	},
}

func fetchProjects() error {
	if time.Now().Sub(lastExecuted(projectsCachePath)) < 24*time.Hour {
		return nil
	}

	projects, err := client.GetProjects(nil)
	if err != nil {
		return err
	}

	base, err := cachePath(projectsCachePath)
	if err != nil {
		return err
	}

	for _, project := range projects {
		data, err := json.Marshal(project)
		if err != nil {
			return err
		}

		path := filepath.Join(base, fmt.Sprintf("%d.json", project.Id))
		if err := ioutil.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}
	if err := setLastExecuted(projectsCachePath); err != nil {
		return err
	}

	return nil
}

func readProjects() (projects []backlog.Project, err error) {
	base, err := cachePath(projectsCachePath)
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

		var project backlog.Project

		if err := json.Unmarshal(data, &project); err != nil {
			return err
		}

		projects = append(projects, project)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func fetchProjectById(projectId uint64) error {
	project, err := client.GetProject(projectId)
	if err != nil {
		return err
	}

	base, err := cachePath(projectsCachePath)
	if err != nil {
		return err
	}

	data, err := json.Marshal(project)
	if err != nil {
		return err
	}

	path := filepath.Join(base, fmt.Sprintf("%d.json", project.Id))
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return err
	}

	return nil
}

func readProjectById(projectId uint64) (project backlog.Project, err error) {
	base, err := cachePath(projectsCachePath)
	if err != nil {
		return project, err
	}

	path := filepath.Join(base, fmt.Sprintf("%d.json", projectId))
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return project, err
	}
	if err := json.Unmarshal(data, &project); err != nil {
		return project, err
	}

	return project, nil
}

func init() {
	projectCommand.AddCommand(projectListCommand)

	rootCommand.AddCommand(projectCommand)
}
