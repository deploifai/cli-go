/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package create

import (
	"context"
	"errors"
	"fmt"
	"github.com/deploifai/cli-go/api"
	"github.com/deploifai/cli-go/api/generated"
	"github.com/deploifai/cli-go/command/ctx"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var cloudProvider generated.CloudProvider

// Cmd represents the create command
var Cmd = &cobra.Command{
	Use:   "create",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1), // requires exactly 1 arg, which is the name of the new cloud profile
	RunE: func(cmd *cobra.Command, args []string) error {
		cloudProfileName := args[0]

		fmt.Println("create called", cloudProfileName, cloudProvider)

		if cloudProvider == "" {
			prompt := promptui.Select{
				Label: "Select cloud provider",
				Items: generated.AllCloudProvider,
			}

			_, result, err := prompt.Run()
			cobra.CheckErr(err)
			cloudProvider = generated.CloudProvider(result)
		} else {
			found := false
			for _, provider := range generated.AllCloudProvider {
				if cloudProvider == provider {
					found = true
					break
				}
			}
			if !found {
				return errors.New(fmt.Sprintf("invalid cloud provider: %s", cloudProvider))
			}
		}

		// todo: remove this once Digital Ocean is supported
		if cloudProvider == generated.CloudProviderDigitalOcean {
			return errors.New("Digital Ocean is not yet supported")
		}

		api := ctx.GetContextValue(cmd).API
		_config := ctx.GetContextValue(cmd).Config

		// check if cloud profile already exists
		// if it does, return an error
		if collision, err := checkCollision(cmd.Context(), *api, _config.Workspace.Username, cloudProfileName, cloudProvider); err != nil {
			cobra.CheckErr(err)
		} else if collision {
			return errors.New(fmt.Sprintf("%s already exists", cloudProfileName))
		}

		// todo: create cloud profile

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

func createCloudProfile(c context.Context, api api.API, username string, cloudProfileName string, provider generated.CloudProvider) (*generated.CreateCloudProfile, error) {
	client := api.GetClient()

	createInput := generated.CloudProfileCreateInput{
		Name:     "test",
		Provider: generated.CloudProviderAws,
	}
	//createInput.AwsCredentials = &generated.AWSCredentials{
	//	AwsAccessKey:       awsAccessKey,
	//	AwsSecretAccessKey: awsSecretKey,
	//}

	data, err := client.CreateCloudProfile(c, username, createInput)

	if err != nil {
		return nil, api.ProcessError(err)
	}

	return data, nil
}
