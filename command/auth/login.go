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
	"encoding/json"
	"errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/deploifai/cli-go/command/ctx"
	"github.com/deploifai/cli-go/host"
	"github.com/spf13/cobra"
	"io"
	"net/http"
)

var token string

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login using a personal access token.",
	RunE: func(cmd *cobra.Command, args []string) error {
		_config := ctx.GetContextValue(cmd).Config

		if _config.Auth.Username != "" && _config.Auth.Token != "" {
			return errors.New("already logged in, try logging out first")
		}

		if token == "" {
			err := survey.AskOne(&survey.Password{
				Message: "Token",
			}, &token)
			cobra.CheckErr(err)
		}

		loginUrl := host.Endpoint.Auth.Login

		request, err := http.NewRequest("POST", loginUrl, bytes.NewBuffer([]byte{}))
		request.Header.Set("Content-Type", "application/json; charset=UTF-8")
		request.Header.Set("authorization", token)
		cobra.CheckErr(err)

		client := &http.Client{}
		response, err := client.Do(request)
		cobra.CheckErr(err)

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			cobra.CheckErr(err)
		}(response.Body)

		if response.StatusCode != 200 {
			return errors.New("invalid token")
		}

		var body struct {
			Username string `json:"username"`
		}
		err = json.NewDecoder(response.Body).Decode(&body)
		cobra.CheckErr(err)

		cmd.Println("Successfully logged in.")

		_config.Auth.Username = body.Username
		_config.Auth.Token = token
		_config.Workspace.Username = body.Username

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
	loginCmd.Flags().StringVarP(&token, "token", "t", "", "generated personal access token")
}
