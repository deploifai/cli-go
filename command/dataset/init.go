/*
Copyright © 2023 Sean Chok
*/
package dataset

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/deploifai/cli-go/command/ctx"
	"github.com/deploifai/sdk-go/api/generated"
	"github.com/deploifai/sdk-go/service/dataset"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [project-name/dataset-name]",
	Short: "Initialise a local directory as a dataset",
	Long: `
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		_context := ctx.GetContextValue(cmd)
		client := dataset.NewFromConfig(*_context.ServiceClientConfig)

		whereAccount := generated.AccountWhereUniqueInput{Username: &_context.Config.Workspace.Username}

		var dataStorage generated.DataStorageFragment

		if len(args) > 0 {
			name := args[0]
			projectName, dataStorageName, err := splitName(name)
			if err != nil {
				return err
			}

			fmt.Println(projectName, dataStorageName)
			dataStorage, err = findDataStorage(cmd.Context(), *client, whereAccount, generated.DataStorageWhereInput{
				Projects: &generated.ProjectListRelationFilter{Some: &generated.ProjectWhereInput{Name: &generated.StringFilter{Equals: &projectName}}},
				Name:     &generated.StringFilter{Equals: &dataStorageName},
			})
		} else {
			dataStorages, err := listDataStorage(cmd.Context(), *client, whereAccount)
			if err != nil {
				return err
			}
			if len(dataStorages) == 0 {
				return errors.New("no dataset found")
			} else {
				dataStorage, err = chooseDataStorage(dataStorages)
			}
		}

		return writeToConfig(dataStorage.GetID())

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
}

func splitName(name string) (string, string, error) {
	split := strings.Split(name, "/")
	if len(split) != 2 {
		return "", "", errors.New("invalid project-name/dataset-name")
	}
	return split[0], split[1], nil
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
		return generated.DataStorageFragment{}, errors.New("no dataset found")
	}

	return data[0], nil
}

func listDataStorage(ctx context.Context, client dataset.Client, whereAccount generated.AccountWhereUniqueInput) ([]generated.DataStorageFragment, error) {

	status := generated.DataStorageStatusDeploySuccess

	dataStorages, err := client.List(ctx, whereAccount, &generated.DataStorageWhereInput{
		Status: &generated.EnumDataStorageStatusFilter{Equals: &status},
	})
	if err != nil {
		return nil, err
	}

	filtered := make([]generated.DataStorageFragment, 0)
	for _, d := range dataStorages {
		if len(d.GetProjects()) > 0 {
			filtered = append(filtered, d)
		}
	}

	return filtered, nil
}

func chooseDataStorage(dataStorages []generated.DataStorageFragment) (generated.DataStorageFragment, error) {

	options := make([]string, len(dataStorages))
	for i, d := range dataStorages {
		options[i] = fmt.Sprintf("%s/%s <%s>", d.GetProjects()[0].GetName(), d.GetName(), d.GetCloudProfile().GetProvider())
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

func writeToConfig(dataStorageId string) error {
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return err
	}

	// write to a config file

	fmt.Println(dataStorageId)
	fmt.Println(currentWorkingDirectory)

	return nil

}
