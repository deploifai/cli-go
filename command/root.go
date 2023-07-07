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
package command

import "C"
import (
	"errors"
	"github.com/deploifai/cli-go/api"
	"github.com/deploifai/cli-go/command/auth"
	"github.com/deploifai/cli-go/command/cloud_profile"
	"github.com/deploifai/cli-go/command/command_config"
	"github.com/deploifai/cli-go/command/ctx"
	"github.com/deploifai/cli-go/command/workspace"
	"github.com/deploifai/cli-go/host"
	"golang.org/x/net/context"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootViper *viper.Viper
var cfgFile string
var rootConfig command_config.Config
var rootAPI api.API

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "deploifai",
	Short: "A CLI for Deploifai",
	Long: `This is a CLI to interact with the Deploifai API. It also provides ` +
		`a lot of powerful tools to super-charge the ML development workflow.`,
	PersistentPostRun: func(cmd *cobra.Command, args []string) {

		rootConfig.WriteStructIntoConfig(rootViper)

		cfgFile := rootViper.ConfigFileUsed()

		// Create config file if it doesn't exist
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Dir(cfgFile), 0755)
			cobra.CheckErr(err)

			err = os.WriteFile(cfgFile, []byte(""), 0600)
			cobra.CheckErr(err)
		}

		// Write config file
		err := rootViper.WriteConfig()
		cobra.CheckErr(err)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add groups of commands
	rootCmd.AddCommand(auth.Cmd, workspace.Cmd, cloud_profile.Cmd)

	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to config file (default to $HOME/.config/deploifai/config.toml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Initialize viper
	rootViper = viper.New()

	// Find home directory.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)
	cfgDirectory := filepath.Join(home, ".config", "deploifai")

	if cfgFile != "" {
		// Use config file from the flag.
		rootViper.SetConfigFile(cfgFile)
	} else {
		// Search config in home/.config/deploifai directory with name "config" (without extension).
		rootViper.AddConfigPath(cfgDirectory)
		rootViper.SetConfigType("toml")
		rootViper.SetConfigName("config")
	}

	// Read in environment variables that match
	rootViper.AutomaticEnv()

	// If a config file is found, read it in.
	err = rootViper.ReadInConfig()
	notFound := &viper.ConfigFileNotFoundError{}

	switch {
	case err != nil && !errors.As(err, notFound):
		cobra.CheckErr(err)
	case err != nil && errors.As(err, notFound):
		// The config file is optional, we shouldn't exit when the config is not found
		rootViper.SetConfigFile(filepath.Join(cfgDirectory, "config.toml"))
		break
	default:
		// No error – do nothing
	}

	// Set defaults
	command_config.SetDefaultConfig(rootViper)

	// Unmarshal config into Struct
	err = rootViper.Unmarshal(&rootConfig)
	cobra.CheckErr(err)

	// Create API
	rootAPI = api.NewAPI(host.Endpoint.GraphQL, host.Endpoint.Base, rootConfig.Auth.Token)

	// Create root command context
	value := ctx.NewContextValue(&rootConfig, &rootAPI)
	_context := context.WithValue(context.Background(), "value", value)
	rootCmd.SetContext(_context)
}
