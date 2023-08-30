/*
Copyright © 2023 Sean Chok
*/
package dataset

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/deploifai/cli-go/command/command_config/project_config"
	"github.com/deploifai/cli-go/command/ctx"
	"github.com/deploifai/sdk-go/api/generated"
	"github.com/deploifai/sdk-go/service/dataset"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var name string

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise the current working directory as a dataset",
	Long: `Initialise the current working directory as a dataset in the current project.

This requires the current working directory or any of its parent directories to be initialised as a project first.
Use the command "deploifai project init" to do that.

If the current working directory or any of its parent directories is already initialised as a dataset, this command will fail.
`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		_context := ctx.GetContextValue(cmd)

		if !_context.Project.Project.IsInitialized() {
			return project_config.ProjectNotInitializedError{}
		}

		// verify if the current directory is already initialised as a dataset
		if ok, _, _, err := getDataset(*_context.Project); err != nil {
			return err
		} else if ok {
			return errors.New("the current directory is already initialised as a dataset")
		}

		projectId := _context.Project.Project.ID

		client := dataset.NewFromConfig(*_context.ServiceClientConfig)
		whereAccount := generated.AccountWhereUniqueInput{Username: &_context.Root.Workspace.Username}

		var dataStorage generated.DataStorageFragment

		if name != "" {
			dataStorage, err = findDataStorage(cmd.Context(), *client, whereAccount, generated.DataStorageWhereInput{
				Projects: &generated.ProjectListRelationFilter{Some: &generated.ProjectWhereInput{ID: &generated.StringFilter{Equals: &projectId}}},
				Name:     &generated.StringFilter{Equals: &name},
			})
			if err != nil {
				return err
			}
		} else {
			dataStorages, err := listDataStorage(cmd.Context(), *client, whereAccount, generated.DataStorageWhereInput{
				Projects: &generated.ProjectListRelationFilter{Some: &generated.ProjectWhereInput{ID: &generated.StringFilter{Equals: &projectId}}},
			})
			if err != nil {
				return err
			}
			if len(dataStorages) == 0 {
				return errors.New("no datasets found in this project")
			} else {
				dataStorage, err = chooseDataStorage(dataStorages)
				if err != nil {
					return err
				}
			}
		}

		return saveInConfig(_context.Project, dataStorage.GetID())

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

	initCmd.Flags().StringVarP(&name, "dataset", "d", "", "name of dataset in the project to use")
}

func findDataStorage(ctx context.Context, client dataset.Client, whereAccount generated.AccountWhereUniqueInput, whereDataStorage generated.DataStorageWhereInput) (generated.DataStorageFragment, error) {

	status := generated.DataStorageStatusDeploySuccess

	data, err := client.List(ctx, whereAccount, &generated.DataStorageWhereInput{
		And: []*generated.DataStorageWhereInput{
			&whereDataStorage,
			{Status: &generated.EnumDataStorageStatusFilter{Equals: &status}},
		},
	})
	if err != nil {
		return generated.DataStorageFragment{}, err
	}

	if len(data) == 0 {
		return generated.DataStorageFragment{}, errors.New(fmt.Sprintf("no dataset found with name %s in this project", name))
	}

	return data[0], nil
}

func listDataStorage(ctx context.Context, client dataset.Client, whereAccount generated.AccountWhereUniqueInput, whereDataStorage generated.DataStorageWhereInput) ([]generated.DataStorageFragment, error) {

	status := generated.DataStorageStatusDeploySuccess

	return client.List(ctx, whereAccount, &generated.DataStorageWhereInput{
		And: []*generated.DataStorageWhereInput{
			&whereDataStorage,
			{Status: &generated.EnumDataStorageStatusFilter{Equals: &status}},
		},
	})
}

func chooseDataStorage(dataStorages []generated.DataStorageFragment) (generated.DataStorageFragment, error) {

	options := make([]string, len(dataStorages))
	for i, d := range dataStorages {
		options[i] = fmt.Sprintf("%s <%s>", d.GetName(), d.GetCloudProfile().GetProvider())
	}

	var index int
	err := survey.AskOne(&survey.Select{
		Message: "Choose a dataset",
		Options: options,
	}, &index, survey.WithPageSize(10))
	if err != nil {
		return generated.DataStorageFragment{}, err
	}

	return dataStorages[index], nil
}

func saveInConfig(projectConfig *project_config.Config, dataStorageId string) error {

	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return err
	}

	relativeDirectory, err := filepath.Rel(filepath.Dir(projectConfig.ConfigFile), currentWorkingDirectory)
	if err != nil {
		return err
	}
	relativeDirectory = filepath.ToSlash(relativeDirectory)

	projectConfig.Datasets[dataStorageId] = project_config.Dataset{
		ID:             dataStorageId,
		LocalDirectory: relativeDirectory,
	}

	return nil
}
