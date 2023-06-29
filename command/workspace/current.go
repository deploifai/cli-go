/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package workspace

import (
	"errors"
	"github.com/deploifai/cli-go/command/ctx"

	"github.com/spf13/cobra"
)

// currentCmd represents the current command
var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the current workspace set.",
	Long: `Show the current workspace set.` +
		` The current workspace is used for all commands that require a workspace.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_config := ctx.GetContextValue(cmd).Config

		if _config.Workspace.Username == "" {
			return errors.New("no workspace set")
		}

		cmd.Println(_config.Workspace.Username)

		return nil
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// currentCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// currentCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
