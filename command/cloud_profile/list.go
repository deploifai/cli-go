/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cloud_profile

import (
	"github.com/deploifai/cli-go/command/ctx"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List cloud profiles in the current workspace.",
	Run: func(cmd *cobra.Command, args []string) {
		_context := ctx.GetContextValue(cmd)
		api := _context.API
		client := api.GetGQLClient()
		_config := _context.Config

		data, err := client.GetCloudProfiles(cmd.Context(), _config.Workspace.Username, nil)
		if err != nil {
			cobra.CheckErr(api.ProcessGQLError(err))
		}

		for i := range data.CloudProfiles {
			cmd.Printf("%s <%s>\n", data.CloudProfiles[i].Name, data.CloudProfiles[i].Provider)
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
