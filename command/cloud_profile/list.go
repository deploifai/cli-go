/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cloud_profile

import (
	"github.com/deploifai/cli-go/command/ctx"
	"github.com/deploifai/sdk-go/api/generated"
	"github.com/deploifai/sdk-go/service/cloud_profile"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List cloud profiles in the current workspace",
	Run: func(cmd *cobra.Command, args []string) {
		_context := ctx.GetContextValue(cmd)
		_config := _context.Root

		client := cloud_profile.NewFromConfig(*_context.ServiceClientConfig)

		cloudProfiles, err := client.List(cmd.Context(), generated.AccountWhereUniqueInput{Username: &_config.Workspace.Username}, nil)
		cobra.CheckErr(err)

		for _, cp := range cloudProfiles {
			cmd.Printf("%s <%s>\n", cp.GetName(), cp.GetProvider())
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
