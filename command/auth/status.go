/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package auth

import (
	"errors"
	"fmt"
	"github.com/deploifai/cli-go/command/ctx"
	"github.com/deploifai/sdk-go/api"
	"github.com/deploifai/sdk-go/api/host"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check login status",
	Run: func(cmd *cobra.Command, args []string) {
		_ctx := ctx.GetContextValue(cmd)
		_config := _ctx.Root

		if _config.Auth.Username == "" || _config.Auth.Token == "" {
			cmd.Println("Not logged in.")
			return // exit
		}

		client := _ctx.ServiceClientConfig.API.GetRestClient()

		checkUri := host.Endpoint.Rest.Auth.Check

		var jsonData = []byte(fmt.Sprintf(`{"username": "%s"}`, _config.Auth.Username))

		request, err := client.NewRequest("POST", checkUri, api.RequestHeaders{
			api.WithContentType(api.ContentTypeJson),
		}, jsonData)
		cobra.CheckErr(err)

		response, err := client.Do(request)
		cobra.CheckErr(err)

		if response.StatusCode == 200 {
			cmd.Println("Logged in as " + _config.Auth.Username + ".")
		} else if response.StatusCode == 401 {
			cmd.Println("Invalid username or token. Please login again.")
		} else {
			cobra.CheckErr(errors.New(fmt.Sprintf("could not check login status with server error status: %s", response.Status)))
		}

	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
