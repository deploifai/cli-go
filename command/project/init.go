/*
Copyright © 2023 Sean Chok
*/
package project

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/deploifai/cli-go/command/command_config/project_config"
	"github.com/deploifai/cli-go/command/ctx"
	"github.com/deploifai/sdk-go/api/generated"
	"github.com/deploifai/sdk-go/service/project"
	"github.com/spf13/cobra"
)

var name string

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the current working directory as a project",
	Long: `Initialize the current working directory as a project in the current workspace.

This creates a deploifai.toml file in the current working directory, which is used to store information related to the project.
This includes datasets under the project, and other project related information.
Do not delete this file, as it is used by the CLI to determine the project context.

If this directory or any of its parent directories is already initialised as a project, this command will fail.
`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		_context := ctx.GetContextValue(cmd)

		if _context.Project.Project.IsInitialized() {
			return errors.New(fmt.Sprintf("%s already exists, this directory is already initialised as a project", project_config.ConfigFilename))
		}

		client := project.NewFromConfig(*_context.ServiceClientConfig)

		var project generated.ProjectFragment

		if name != "" {
			if project, err = findProject(cmd.Context(), *client, _context.Root.Workspace.Username, name); err != nil {
				return err
			}
		} else {
			project, err = chooseProject(cmd.Context(), *client, _context.Root.Workspace.Username)
			if err != nil {
				return err
			}
		}

		// save in project config
		_context.Project.Project.ID = project.ID

		return nil
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	initCmd.Flags().StringVarP(&name, "project", "p", "", "name of project in the current workspace to use")
}

func findProject(ctx context.Context, client project.Client, username string, projectName string) (project generated.ProjectFragment, err error) {

	status := generated.ProjectStatusSetupSuccess

	projects, err := client.List(ctx, generated.AccountWhereUniqueInput{Username: &username}, &generated.ProjectWhereInput{
		Name:   &generated.StringFilter{Equals: &projectName},
		Status: &generated.EnumProjectStatusFilter{Equals: &status},
	})

	if err != nil {
		return project, err
	}

	if len(projects) == 0 {
		return project, errors.New(fmt.Sprintf("project with name: %s not found", projectName))
	}

	return projects[0], nil
}

func chooseProject(ctx context.Context, client project.Client, username string) (project generated.ProjectFragment, err error) {

	projects, err := client.List(ctx, generated.AccountWhereUniqueInput{Username: &username}, nil)
	if err != nil {
		return project, err
	}

	if len(projects) == 0 {
		return project, errors.New("no project found")
	}

	options := make([]string, len(projects))
	for i, project := range projects {
		options[i] = project.Name
	}

	var index int
	err = survey.AskOne(&survey.Select{
		Message: "Choose a project",
		Options: options,
	}, &index, survey.WithPageSize(10))
	if err != nil {
		return project, err
	}

	return projects[index], nil

}
