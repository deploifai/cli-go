/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package auth

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/deploifai/cli-go/command/command_config"
	"net/http"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check login status.",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		_config := command_config.GetConfig(cmd)

		if _config.Auth.Username == "" || _config.Auth.Token == "" {
			cmd.Println("Not logged in.")
		}

		loginUrl := backendUrl + "/auth/login/cli"

		var jsonData = []byte(fmt.Sprintf(`{"username": "%s"}`, _config.Auth.Username))

		request, err := http.NewRequest("POST", loginUrl, bytes.NewBuffer(jsonData))
		request.Header.Set("Content-Type", "application/json; charset=UTF-8")
		request.Header.Set("authorization", _config.Auth.Token)
		cobra.CheckErr(err)

		client := &http.Client{}
		response, err := client.Do(request)
		cobra.CheckErr(err)

		if response.StatusCode == 200 {
			cmd.Println("Logged in as " + _config.Auth.Username + ".")
		} else if response.StatusCode == 401 {
			cmd.Println("Invalid username or token. Please login again.")
		} else {
			cobra.CheckErr(errors.New("could not check login status with server"))
		}

	},
}

func init() {
	Cmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
