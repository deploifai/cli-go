/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package workspace

import (
	"github.com/spf13/cobra"
)

// Cmd represents the workspace command
var Cmd = &cobra.Command{
	Use:     "workspace",
	Aliases: []string{"ws"},
	Short:   "Manage workspaces",
	Long:    `List workspaces, or set, and show the current workspace.`,
}

func init() {
	Cmd.AddCommand(currentCmd, setCmd, listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// workspaceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// workspaceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
