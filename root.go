package main

import (
	"fmt"
	"os"

	backlog "github.com/moutend/go-backlog"
	"github.com/spf13/cobra"
)

var (
	debug  bool
	space  string
	token  string
	client *backlog.Client
)

var rootCommand = &cobra.Command{
	Use: "backlog",
	RunE: func(c *cobra.Command, args []string) error {
		var err error

		space = os.Getenv("BACKLOG_SPACE")
		token = os.Getenv("BACKLOG_TOKEN")

		client, err = backlog.New(space, token)
		if err != nil {
			return err
		}

		return nil
	},
	PersistentPostRunE: func(c *cobra.Command, args []string) error {
		command := c
		commands := []string{}

		for {
			commands = append(commands, command.Use)

			if command = command.Parent(); command == nil {
				break
			}
		}

		fmt.Println("@@@", commands)
		return nil
	},
}

func init() {
	rootCommand.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug enable flag")
}
