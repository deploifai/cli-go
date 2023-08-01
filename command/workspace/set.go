/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package workspace

import (
	"errors"
	"github.com/deploifai/cli-go/command/ctx"
	"github.com/deploifai/sdk-go/service/workspace"

	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set <workspace>",
	Short: "Set the current workspace used",
	Long: `Set the current workspace used.` +
		` The current workspace is used for all commands that require a workspace.`,
	Args: cobra.ExactArgs(1), // require exactly 1 arg
	RunE: func(cmd *cobra.Command, args []string) error {
		newWorkspace := args[0] // first arg

		// verify workspace exists
		cfg := ctx.GetContextValue(cmd).ServiceClientConfig
		client := workspace.NewFromConfig(*cfg)

		workspaces, err := client.List(cmd.Context())
		cobra.CheckErr(err)

		var found bool
		for _, workspace := range workspaces {
			if newWorkspace == workspace.GetUsername() {
				found = true
				break
			}
		}

		if !found {
			return errors.New("workspace not found")
		}

		// save new workspace to config
		_config := ctx.GetContextValue(cmd).Config
		_config.Workspace.Username = newWorkspace

		cmd.Println(newWorkspace)

		return nil
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
