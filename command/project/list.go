/*
Copyright © 2023 Sean Chok
*/
package project

import (
	"github.com/deploifai/cli-go/command/ctx"
	"github.com/deploifai/sdk-go/api/generated"
	"github.com/deploifai/sdk-go/service/cloud_profile"
	"github.com/deploifai/sdk-go/service/project"
	"github.com/spf13/cobra"
)

type projectInList struct {
	name             string
	cloudProfileName string
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects in the current workspace",
	RunE: func(cmd *cobra.Command, args []string) error {

		_context := ctx.GetContextValue(cmd)
		_config := _context.Root

		projectClient := project.NewFromConfig(*_context.ServiceClientConfig)

		projects, err := projectClient.List(cmd.Context(), generated.AccountWhereUniqueInput{Username: &_config.Workspace.Username}, nil)
		if err != nil {
			return err
		}

		cloudProfileClient := cloud_profile.NewFromConfig(*_context.ServiceClientConfig)

		projectList := make([]projectInList, len(projects))

		for i, p := range projects {
			cp, err := cloudProfileClient.Get(cmd.Context(), generated.CloudProfileWhereUniqueInput{ID: p.GetCloudProfileID()})
			if err != nil {
				return err
			}

			projectList[i] = projectInList{
				name:             p.GetName(),
				cloudProfileName: cp.GetName(),
			}
		}

		for _, p := range projectList {
			cmd.Printf("%s <%s>\n", p.name, p.cloudProfileName)
		}

		return nil
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
