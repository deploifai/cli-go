/*
Copyright © 2023 Sean Chok <seanchok@deploif.ai>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package auth

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/deploifai/cli-go/command/command_config"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"net/http"
)

const backendUrl = "https://api.deploif.ai"

var username string
var token string

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login using a username and a personal access token.",
	RunE: func(cmd *cobra.Command, args []string) error {
		_config := command_config.GetConfig(cmd)

		if _config.Auth.Username != "" && _config.Auth.Token != "" {
			return errors.New("already logged in, try logging out first")
		}

		if username == "" {
			prompt := promptui.Prompt{
				Label: "username",
				Validate: func(input string) error {
					if len(input) < 1 {
						return errors.New("username cannot be empty")
					}
					return nil
				},
			}
			result, err := prompt.Run()
			cobra.CheckErr(err)
			username = result
		}

		if token == "" {
			prompt := promptui.Prompt{
				Label: "token",
				Validate: func(input string) error {
					if len(input) < 1 {
						return errors.New("username cannot be empty")
					}
					return nil
				},
			}
			result, err := prompt.Run()
			cobra.CheckErr(err)
			token = result
		}

		loginUrl := backendUrl + "/auth/login/cli"

		var jsonData = []byte(fmt.Sprintf(`{"username": "%s"}`, username))

		request, err := http.NewRequest("POST", loginUrl, bytes.NewBuffer(jsonData))
		request.Header.Set("Content-Type", "application/json; charset=UTF-8")
		request.Header.Set("authorization", token)
		cobra.CheckErr(err)

		client := &http.Client{}
		response, err := client.Do(request)
		cobra.CheckErr(err)

		if response.StatusCode != 200 {
			return errors.New("invalid username or token")
		}

		cmd.Println("Successfully logged in.")

		_config.Auth.Username = username
		_config.Auth.Token = token
		_config.Workspace.Username = username

		return nil
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	loginCmd.Flags().StringVarP(&username, "username", "u", "", "Deploifai username")
	loginCmd.Flags().StringVarP(&token, "token", "t", "", "generated personal access token")
}