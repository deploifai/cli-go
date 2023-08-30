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

import (
	"errors"
	"fmt"
	"github.com/deploifai/cli-go/command/auth"
	"github.com/deploifai/cli-go/command/cloud_profile"
	"github.com/deploifai/cli-go/command/command_config/project_config"
	"github.com/deploifai/cli-go/command/command_config/root_config"
	"github.com/deploifai/cli-go/command/ctx"
	"github.com/deploifai/cli-go/command/dataset"
	"github.com/deploifai/cli-go/command/project"
	"github.com/deploifai/cli-go/command/workspace"
	"github.com/deploifai/sdk-go/config"
	"github.com/deploifai/sdk-go/credentials"
	"golang.org/x/net/context"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootViper *viper.Viper
var rootConfigFile string
var rootConfig root_config.Config
var rootServiceClientConfig config.Config

var projectViper *viper.Viper
var projectConfig project_config.Config

var configFileNotFoundErr = &viper.ConfigFileNotFoundError{}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "deploifai",
	Short: "A CLI for Deploifai",
	Long: `This is a CLI to interact with the Deploifai API. It also provides ` +
		`a lot of powerful tools to super-charge the ML development workflow.`,
	PersistentPostRun: func(cmd *cobra.Command, args []string) {

		rootConfig.WriteStructIntoViper(rootViper)
		err := writeConfig(rootViper, 0755, 0600)
		cobra.CheckErr(err)

		// if the project in project config is not empty
		// write to file
		if projectConfig.Project.ID != "" {
			projectConfig.WriteStructIntoViper(projectViper)
			err = writeConfig(projectViper, 0755, 0644)
			cobra.CheckErr(err)
		}

	},
}

func init() {
	// Add groups of commands
	rootCmd.AddCommand(versionCmd, auth.Cmd, workspace.Cmd, cloud_profile.Cmd, project.Cmd, dataset.Cmd)

	cobra.OnInitialize(initConfigs)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	defaultConfigFile := filepath.Join("$HOME", ".config", "deploifai", "config.toml")
	rootCmd.PersistentFlags().StringVar(&rootConfigFile, "config", "", fmt.Sprintf("path to config file (default to %s)", defaultConfigFile))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfigs reads config files and ENV variables if set.
func initConfigs() {

	bgCtx := context.Background()

	// Initialize root config
	err := initRootConfig()
	cobra.CheckErr(err)

	// Initialize project config
	err = initProjectConfig()
	cobra.CheckErr(err)

	// Create service client config
	rootServiceClientConfig, err = config.LoadDefaultConfig(bgCtx, config.WithCredentials(credentials.NewCredentials(rootConfig.Auth.Token)))
	cobra.CheckErr(err)

	// Create root command context
	value := ctx.NewContextValue(&rootConfig, &projectConfig, &rootServiceClientConfig)
	_context := context.WithValue(bgCtx, "value", value)
	rootCmd.SetContext(_context)
}

func initRootConfig() error {

	// Initialize viper
	rootViper = viper.New()

	// Find home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	cfgDirectory := filepath.Join(home, ".config", "deploifai")

	if rootConfigFile != "" {
		// Use config file from the flag
		rootViper.SetConfigFile(rootConfigFile)
	} else {
		// Search config in home/.config/deploifai directory with name "config" (without extension)
		rootViper.AddConfigPath(cfgDirectory)
		rootViper.SetConfigType("toml")
		rootViper.SetConfigName("config")
	}

	// Read in environment variables that match
	rootViper.AutomaticEnv()

	// If a config file is found, read it in
	err = rootViper.ReadInConfig()

	switch {
	case err != nil && !errors.As(err, configFileNotFoundErr):
		return err
	case err != nil && errors.As(err, configFileNotFoundErr):
		// The config file is optional, we shouldn't exit when the config is not found
		rootViper.SetConfigFile(filepath.Join(cfgDirectory, "config.toml"))
		break
	default:
		// No error – do nothing
	}

	// Set defaults
	root_config.SetDefaultConfig(rootViper)

	// Unmarshal config into Struct
	return rootViper.Unmarshal(&rootConfig)

}

func initProjectConfig() error {

	// Initialize viper
	projectViper = viper.New()

	// Find project config file
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	dir, err := findProjectConfigDir(cwd, project_config.ConfigFilename)
	if err != nil {
		return err
	}

	if dir != "" {
		projectViper.SetConfigFile(filepath.Join(dir, project_config.ConfigFilename))
	}

	err = projectViper.ReadInConfig()

	switch {
	case err != nil && !errors.As(err, configFileNotFoundErr):
		return err
	case err != nil && errors.As(err, configFileNotFoundErr):
		// The config file is optional, we shouldn't exit when the config is not found
		projectViper.SetConfigFile(filepath.Join(cwd, project_config.ConfigFilename))
		break
	default:
		// No error – do nothing
	}

	// Set defaults
	project_config.SetDefaultConfig(projectViper)

	// Unmarshal config into Struct
	if err = projectViper.Unmarshal(&projectConfig); err != nil {
		return err
	}

	// Set config file
	projectConfig.SetConfigFile(projectViper.ConfigFileUsed())

	return nil

}

func findProjectConfigDir(dir string, configFilename string) (string, error) {

	f, err := os.Stat(filepath.Join(dir, configFilename))

	if err == nil {
		if f.IsDir() {
			return "", errors.New(fmt.Sprintf("%s already exists as a directory, this is not allowed as %s should be a file", configFilename, configFilename))
		}
		return dir, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		// file does not exist

		// check if dir is root
		if dir == filepath.Dir(dir) {
			return "", nil
		}

		// check parent dir
		return findProjectConfigDir(filepath.Dir(dir), configFilename)
	}

	return "", err
}

func writeConfig(v *viper.Viper, dirPerm os.FileMode, filePerm os.FileMode) error {

	cfgFile := v.ConfigFileUsed()

	// Create config file if it doesn't exist
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(cfgFile), dirPerm); err != nil {
			return err
		}

		if err = os.WriteFile(cfgFile, []byte(""), filePerm); err != nil {
			return err
		}
	}

	// Write config file
	return v.WriteConfig()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
