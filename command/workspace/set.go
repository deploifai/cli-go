/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package workspace

import (
	"errors"
	"github.com/deploifai/cli-go/command/ctx"

	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set <workspace>",
	Short: "Set the current workspace used.",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1), // require exactly 1 arg
	RunE: func(cmd *cobra.Command, args []string) error {
		newWorkspace := args[0] // first arg

		// verify workspace exists
		api := ctx.GetContextValue(cmd).API
		client := api.GetClient()
		data, err := client.GetAccounts(cmd.Context())
		if err != nil {
			cobra.CheckErr(api.ProcessError(err))
		}

		found := newWorkspace == data.Me.Account.Username
		if !found {
			for i := range data.Me.Teams {
				if newWorkspace == data.Me.Teams[i].Account.Username {
					found = true
					break
				}
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
