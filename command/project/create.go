/*
Copyright © 2023 Sean Chok
*/
package project

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/deploifai/cli-go/command/ctx"
	"github.com/deploifai/cli-go/utils/spinner_utils"
	"github.com/deploifai/sdk-go/api/generated"
	"github.com/deploifai/sdk-go/service/cloud_profile"
	"github.com/deploifai/sdk-go/service/project"
	"github.com/spf13/cobra"
)

var notDefaultCloudProfile = false

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new project in the current workspace",
	Long: `Create a new project on Deploifai for a new Machine Learning project.

Every project requires a cloud profile to first create a project-specific dataset that can be used to store the results of experiments.
`,
	Args: cobra.ExactArgs(1), // requires exactly 1 arg, which is the name of the new cloud profile
	RunE: func(cmd *cobra.Command, args []string) error {

		projectName := args[0]

		_context := ctx.GetContextValue(cmd)

		cloudProfileClient := cloud_profile.NewFromConfig(*_context.ServiceClientConfig)
		projectClient := project.NewFromConfig(*_context.ServiceClientConfig)
		whereAccount := generated.AccountWhereUniqueInput{Username: &_context.Config.Workspace.Username}

		cloudProfile, err := getCloudProfile(cmd.Context(), *cloudProfileClient, whereAccount)
		if err != nil {
			return err
		}

		// check if project name already exists
		if collision, err := checkCollision(cmd.Context(), *projectClient, whereAccount, projectName); err != nil {
			cobra.CheckErr(err)
		} else if collision {
			return errors.New(fmt.Sprintf("%s already exists in the current workspace", projectName))
		}

		// create project
		newProject, err := createProject(cmd.Context(), *projectClient, whereAccount, generated.CreateProjectInput{
			Name:           projectName,
			CloudProfileID: cloudProfile.GetID()},
		)

		cmd.Printf("Successfully created project %s\n", newProject.GetName())

		return nil
	},
}

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	createCmd.Flags().BoolVar(&notDefaultCloudProfile, "not-default-cp", false, "select a cloud profile instead of using the default cloud profile in this workspace to create a new project")
}

func getCloudProfile(c context.Context, client cloud_profile.Client, whereAccount generated.AccountWhereUniqueInput) (generated.CloudProfileFragment, error) {
	if notDefaultCloudProfile {
		return chooseCloudProfile(c, client, whereAccount)
	} else {
		return getDefaultCloudProfile(c, client, whereAccount)
	}
}

func getDefaultCloudProfile(c context.Context, client cloud_profile.Client, whereAccount generated.AccountWhereUniqueInput) (generated.CloudProfileFragment, error) {

	cloudProfiles, err := client.List(c,
		whereAccount,
		&generated.CloudProfileWhereInput{
			DefaultAccount: &generated.AccountRelationFilter{
				Is: &generated.AccountWhereInput{
					Username: &generated.StringFilter{
						Equals: whereAccount.Username}}}},
	)
	if err != nil {
		return generated.CloudProfileFragment{}, err
	}

	if len(cloudProfiles) == 0 {
		return generated.CloudProfileFragment{}, errors.New("no cloud profiles found in the current workspace, please create one first")
	}

	return cloudProfiles[0], nil
}

func chooseCloudProfile(c context.Context, client cloud_profile.Client, whereAccount generated.AccountWhereUniqueInput) (generated.CloudProfileFragment, error) {
	cloudProfiles, err := client.List(c, whereAccount, nil)
	if err != nil {
		return generated.CloudProfileFragment{}, err
	}

	if len(cloudProfiles) == 0 {
		return generated.CloudProfileFragment{}, errors.New("no cloud profiles found in the current workspace, please create one first")
	}

	options := make([]string, len(cloudProfiles))
	for i, cp := range cloudProfiles {
		options[i] = fmt.Sprintf("%s <%s>", cp.GetName(), cp.GetProvider())
	}

	var index int
	err = survey.AskOne(&survey.Select{
		Message: "Choose a cloud profile",
		Options: options,
	}, &index, survey.WithPageSize(10))
	if err != nil {
		return generated.CloudProfileFragment{}, err
	}

	return cloudProfiles[index], nil
}

func checkCollision(c context.Context, client project.Client, whereAccount generated.AccountWhereUniqueInput, projectName string) (bool, error) {

	projects, err := client.List(c, whereAccount, &generated.ProjectWhereInput{
		Name: &generated.StringFilter{
			Equals: &projectName,
		},
	})

	if err != nil {
		return false, err
	}

	return len(projects) > 0, nil
}

func createProject(c context.Context, client project.Client, whereAccount generated.AccountWhereUniqueInput, input generated.CreateProjectInput) (generated.ProjectFragment, error) {

	spinner := spinner_utils.NewAPICallSpinner()
	spinner.Suffix = " Creating project... "

	spinner.Start()
	defer spinner.Stop()

	return client.Create(c, whereAccount, input)
}
