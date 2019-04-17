package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	backlog "github.com/moutend/go-backlog"
	"github.com/spf13/cobra"
)

var wikiCommand = &cobra.Command{
	Use: "wiki",
	RunE: func(c *cobra.Command, args []string) error {
		if err := rootCommand.RunE(c, args); err != nil {
			return err
		}

		path, err := cachePath(wikisCachePath)
		if err != nil {
			return err
		}

		os.MkdirAll(path, 0755)

		return nil
	},
}

var wikiListCommand = &cobra.Command{
	Use: "list",
	RunE: func(c *cobra.Command, args []string) error {
		if err := wikiCommand.RunE(c, args); err != nil {
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
			query := url.Values{}
			query.Add("projectIdOrKey", fmt.Sprint(project.Id))
			if err := fetchWikis(query); err != nil {
				return err
			}

			wikis, err := readWikis(project.Id)
			if err != nil {
				return err
			}

			sort.Slice(wikis, func(i, j int) bool {
				return wikis[i].Updated.Time().After(wikis[j].Updated.Time())
			})

			fmt.Printf("- [%s] %s\n", project.ProjectKey, project.Name)

			for _, wiki := range wikis {
				fmt.Printf("  - %s updated %s (%d)\n", wiki.Name, wiki.Updated.Time().Format("2006-01-02"), wiki.Id)
			}
		}

		return nil
	},
}

var wikiShowCommand = &cobra.Command{
	Use: "show",
	RunE: func(c *cobra.Command, args []string) error {
		if err := wikiCommand.RunE(c, args); err != nil {
			return err
		}
		if len(args) == 0 {
			return nil
		}

		wikiId, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		if err := fetchWiki(uint64(wikiId)); err != nil {
			return err
		}

		wiki, err := readWiki(uint64(wikiId))
		if err != nil {
			return err
		}

		if err := fetchProjectById(wiki.ProjectId); err != nil {
			return err
		}

		project, err := readProjectById(wiki.ProjectId)
		if err != nil {
			return err
		}

		{
			fmt.Println("---")
			fmt.Println("project:", project.ProjectKey)
			fmt.Printf("name: %s\n", wiki.Name)
			fmt.Printf("created: %s\n", wiki.Created.Time().Format("2006-01-02"))
			fmt.Printf("updated: %s\n", wiki.Updated.Time().Format("2006-01-02"))

			u, err := url.Parse(fmt.Sprintf("https://%s.backlog.jp/wiki/%s/%s", space, project.ProjectKey, wiki.Name))
			if err != nil {
				return err
			}

			fmt.Println("url:", u)
			fmt.Println("---")
			fmt.Println(wiki.Content)
		}

		return nil
	},
}

func fetchWikis(query url.Values) error {
	if time.Now().Sub(lastExecuted(wikisCachePath, query)) < 30*time.Minute {
		return nil
	}

	wikis, err := client.GetWikis(query)
	if err != nil {
		return err
	}

	base, err := cachePath(wikisCachePath)
	if err != nil {
		return err
	}

	for _, wiki := range wikis {
		data, err := json.Marshal(wiki)
		if err != nil {
			return err
		}

		path := filepath.Join(base, fmt.Sprintf("%d.json", wiki.Id))
		if err := ioutil.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}
	if err := setLastExecuted(wikisCachePath, query); err != nil {
		return err
	}

	return nil
}

func fetchWiki(wikiId uint64) error {
	if time.Now().Sub(lastExecuted(wikiCachePath, nil)) < 30*time.Minute {
		return nil
	}

	wiki, err := client.GetWiki(wikiId)
	if err != nil {
		return err
	}

	base, err := cachePath(wikisCachePath)
	if err != nil {
		return err
	}

	data, err := json.Marshal(wiki)
	if err != nil {
		return err
	}

	path := filepath.Join(base, fmt.Sprintf("%d.json", wiki.Id))
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return err
	}
	if err := setLastExecuted(wikiCachePath, nil); err != nil {
		return err
	}

	return nil
}

func readWikis(projectId uint64) (wikis []backlog.Wiki, err error) {
	base, err := cachePath(wikisCachePath)
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

		var wiki backlog.Wiki

		if err := json.Unmarshal(data, &wiki); err != nil {
			return err
		}
		if wiki.ProjectId == projectId {
			wikis = append(wikis, wiki)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return wikis, nil
}

func readWiki(wikiId uint64) (wiki backlog.Wiki, err error) {
	base, err := cachePath(wikisCachePath)
	if err != nil {
		return wiki, err
	}

	path := filepath.Join(base, fmt.Sprintf("%d.json", wikiId))
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return wiki, err
	}
	if err := json.Unmarshal(data, &wiki); err != nil {
		return wiki, err
	}

	return wiki, nil
}

func init() {
	wikiCommand.AddCommand(wikiListCommand)
	wikiCommand.AddCommand(wikiShowCommand)

	rootCommand.AddCommand(wikiCommand)
}
