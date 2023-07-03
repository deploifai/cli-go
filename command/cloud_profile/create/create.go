/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package create

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/deploifai/cli-go/api"
	"github.com/deploifai/cli-go/api/generated"
	"github.com/deploifai/cli-go/api/utils"
	"github.com/deploifai/cli-go/command/ctx"
	"github.com/deploifai/cli-go/utils/spinner_utils"
	"github.com/spf13/cobra"
)

var cloudProvider generated.CloudProvider

// Cmd represents the create command
var Cmd = &cobra.Command{
	Use:   "create",
	Short: "Create a cloud profile in the current workspace.",
	Long: `Create cloud credentials for a cloud provider to be used to provision resources in the current workspace.

Currently supported cloud providers:
- AWS
- Azure
- GCP
`,
	Args: cobra.ExactArgs(1), // requires exactly 1 arg, which is the name of the new cloud profile
	RunE: func(cmd *cobra.Command, args []string) error {
		cloudProfileName := args[0]

		if cloudProvider == "" {
			options := make([]string, len(generated.AllCloudProvider))
			for i, provider := range generated.AllCloudProvider {
				options[i] = string(provider)
			}

			answer := ""

			err := survey.AskOne(&survey.Select{
				Message: "Select cloud provider",
				Options: options,
			}, &answer, survey.WithValidator(survey.Required))

			cobra.CheckErr(err)
			cloudProvider = generated.CloudProvider(answer)
		} else {
			found := false
			for _, provider := range generated.AllCloudProvider {
				if cloudProvider == provider {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("invalid cloud provider: %s", cloudProvider)
			}
		}

		// todo: remove this once Digital Ocean is supported
		if cloudProvider == generated.CloudProviderDigitalOcean {
			return errors.New("Digital Ocean is not supported yet")
		}

		api := ctx.GetContextValue(cmd).API
		_config := ctx.GetContextValue(cmd).Config

		// check if cloud profile already exists
		// if it does, return an error
		if collision, err := checkCollision(cmd.Context(), *api, _config.Workspace.Username, cloudProfileName, cloudProvider); err != nil {
			cobra.CheckErr(err)
		} else if collision {
			return errors.New(fmt.Sprintf("%s for %s already exists in the current workspace", cloudProfileName, cloudProvider))
		}

		createInput, err := createCredentialsOnProvider(cmd.Context(), cloudProfileName, cloudProvider)
		if err != nil {
			cobra.CheckErr(err)
		}

		cloudProfile, err := createCloudProfile(cmd.Context(), *api, _config.Workspace.Username, createInput)
		if err != nil {
			cobra.CheckErr(err)
		}

		cmd.Printf("Successfully created cloud profile %s\n", cloudProfile.Name)

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
	Cmd.Flags().StringVarP((*string)(&cloudProvider), "provider", "p", "", "cloud provider, must be one of: AWS, AZURE, GCP")
}

func checkCollision(c context.Context, api api.API, username string, cloudProfileName string, provider generated.CloudProvider) (bool, error) {
	client := api.GetClient()

	data, err := client.GetCloudProfiles(c, username, &generated.CloudProfileWhereInput{
		Name: &generated.StringFilter{
			Equals: &cloudProfileName,
		},
		Provider: &generated.EnumCloudProviderFilter{
			Equals: &provider,
		},
	})

	if err != nil {
		return false, api.ProcessError(err)
	}

	return len(data.CloudProfiles) > 0, nil
}

func createCredentialsOnProvider(ctx context.Context, name string, provider generated.CloudProvider) (createInput generated.CloudProfileCreateInput, err error) {

	createInput = generated.CloudProfileCreateInput{
		Name:     name,
		Provider: provider,
	}

	credentialsCreatorWrapper := NewCredentialsCreatorWrapper(ctx, provider)

	profiles, err := credentialsCreatorWrapper.credentialsCreator.getProfiles()
	if err != nil {
		return createInput, err
	}

	profile, err := credentialsCreatorWrapper.promptProfile(profiles)
	if err != nil {
		return createInput, err
	}

	credentials, err := credentialsCreatorWrapper.credentialsCreator.createCredentials(profile, name)
	if err != nil {
		return createInput, err
	}

	credentialsCreatorWrapper.populateInput(&createInput, credentials)

	return createInput, err
}

func createCloudProfile(ctx context.Context, _api api.API, username string, input generated.CloudProfileCreateInput) (*generated.CreateCloudProfile_CreateCloudProfile, error) {

	spinner := spinner_utils.NewAPICallSpinner()
	spinner.Suffix = " Creating cloud profile... "

	client := _api.GetClient()

	spinner.Start()
	defer spinner.Stop()

	f := utils.CallWithRetries[*generated.CreateCloudProfile]
	data, err := f(func() (*generated.CreateCloudProfile, error) {
		data, err := client.CreateCloudProfile(ctx, username, input)
		return data, err
	}, 10)

	if err != nil {
		return nil, _api.ProcessError(err)
	}

	return data.GetCreateCloudProfile(), nil
}
