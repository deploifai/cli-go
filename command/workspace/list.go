/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package workspace

import (
	"github.com/deploifai/cli-go/command/ctx"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		api := ctx.GetContextValue(cmd).API
		client := api.GetClient()

		data, err := client.GetAccounts(cmd.Context())
		if err != nil {
			cobra.CheckErr(api.ProcessError(err))
		}

		_config := ctx.GetContextValue(cmd).Config

		// print personal workspace
		cmd.Println(_config.Workspace.Username, "<Personal>")

		for i := range data.Me.Teams {
			cmd.Println(data.Me.Teams[i].Account.Username, "<Team>")
		}
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
