/*
Copyright © 2023 Sean Chok
*/
package dataset

import (
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push [<path>...]",
	Short: "Push local files to a dataset",
	Long: `Upload local files to a dataset.

This requires the local directory to be initialised as a dataset first.
Use the command "deploifai dataset init" to do that.

If no path is specified, the current directory is used.
`,
	RunE: func(cmd *cobra.Command, args []string) error {

		// get the dataset directory path from config
		datasetDirPath, err := getDatasetDirPath()

		// get the srcAbsPaths from args
		srcAbsPaths, err := getSrcAbsPaths(args)
		if err != nil {
			return err
		}

		// get the remoteObjectPrefixes from srcAbsPaths
		remoteObjectPrefixes, err := getRemoteObjectPrefixes(datasetDirPath, srcAbsPaths)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pushCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pushCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getDatasetDirPath() (string, error) {

	// todo: look for dataset dir path in config

	return ".", nil
}

func getSrcAbsPaths(relativePaths []string) ([]string, error) {
	if len(relativePaths) == 0 {
		currentWorkingDirectory, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		return []string{currentWorkingDirectory}, nil
	} else {
		return convertToSrcAbsPaths(relativePaths)
	}
}

func convertToSrcAbsPaths(paths []string) (targets []string, err error) {
	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		targets = append(targets, absPath)
	}

	return targets, nil
}

func getRemoteObjectPrefixes(datasetDirPath string, srcAbsPaths []string) (remoteObjectPrefix []string, err error) {
	for _, srcAbsPath := range srcAbsPaths {
		relativePath, err := filepath.Rel(datasetDirPath, srcAbsPath)
		if err != nil {
			return nil, err
		}
		remoteObjectPrefix = append(remoteObjectPrefix, relativePath)
	}

	return remoteObjectPrefix, nil
}
