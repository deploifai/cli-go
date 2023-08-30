/*
Copyright © 2023 Sean Chok
*/
package dataset

import (
	"context"
	"errors"
	"fmt"
	"github.com/deploifai/cli-go/command/command_config/project_config"
	"github.com/deploifai/cli-go/command/ctx"
	"github.com/deploifai/cli-go/utils/spinner_utils"
	"github.com/deploifai/sdk-go/api/generated"
	"github.com/deploifai/sdk-go/service/dataset"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
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

		_context := ctx.GetContextValue(cmd)

		// get the ds directory path from config
		ds, datasetDirPath, err := getDataset(*_context.Project)
		if err != nil {
			return err
		}

		// get the srcAbsPaths from args
		srcAbsPaths, err := getSrcAbsPaths(args)
		if err != nil {
			return err
		}

		ok, invalidArgs, err := verifyPaths(datasetDirPath, args, srcAbsPaths)
		if err != nil {
			return err
		} else if !ok {
			return errors.New(fmt.Sprintf("invalid paths: %s", strings.Join(invalidArgs, ", ")))
		}

		fmt.Println("ds: ", ds)
		fmt.Println("datasetDirPath: ", datasetDirPath)

		// get the remoteObjectPrefixes from srcAbsPaths
		remoteObjectPrefixes, err := getRemoteObjectPrefixes(datasetDirPath, srcAbsPaths)
		if err != nil {
			return err
		}

		fmt.Println("srcAbsPaths: ", srcAbsPaths)
		fmt.Println("remoteObjectPrefixes: ", remoteObjectPrefixes)

		client := dataset.NewFromConfig(*_context.ServiceClientConfig)

		for i, path := range srcAbsPaths {
			srcRelPath := "."
			if len(args) > 0 {
				srcRelPath = args[i]
			}
			if err = push(cmd.Context(), *client, ds.ID, srcRelPath, path, remoteObjectPrefixes[i]); err != nil {
				return err
			}
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

func isSubDir(parent string, child string) (bool, error) {

	up := filepath.Join("..", string(filepath.Separator))

	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false, err
	}

	return !strings.HasPrefix(rel, up) && rel != "..", nil

}

func verifyPaths(datasetDirPath string, args []string, paths []string) (ok bool, invalidArgs []string, err error) {
	for i, path := range paths {
		// check if path exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			invalidArgs = append(invalidArgs, args[i])
		} else if err != nil {
			return false, nil, err
		}

		// check if path is a subdirectory of datasetDirPath
		if ok, err := isSubDir(datasetDirPath, path); err != nil {
			return false, nil, err
		} else if !ok {
			invalidArgs = append(invalidArgs, args[i])
		}
	}

	if len(invalidArgs) > 0 {
		return false, invalidArgs, nil
	} else {
		return true, nil, nil
	}
}

func getDataset(projectConfig project_config.Config) (dataset project_config.Dataset, dirPath string, err error) {

	cwd, err := os.Getwd()
	if err != nil {
		return dataset, dirPath, err
	}

	projectDir := filepath.Dir(projectConfig.ConfigFile)

	for _, d := range projectConfig.Datasets {
		datasetDirPath := filepath.Join(projectDir, filepath.FromSlash(d.LocalDirectory))

		ok, err := isSubDir(datasetDirPath, cwd)
		if err != nil {
			return dataset, dirPath, err
		} else if ok {
			return d, datasetDirPath, nil
		}
	}

	return dataset, dirPath, errors.New("no dataset found, the current directory is not initialised as a dataset")
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

func push(ctx context.Context, client dataset.Client, dataStorageId string, srcRelPath string, srcAbsPath string, remoteObjectPrefix string) error {

	fileInfo, err := os.Stat(srcAbsPath)
	if err != nil {
		return err
	}

	whereDataStorage := generated.DataStorageWhereUniqueInput{ID: &dataStorageId}

	if fileInfo.IsDir() {
		// upload directory
		if err = pushDir(ctx, client, whereDataStorage, srcRelPath, srcAbsPath, remoteObjectPrefix); err != nil {
			return err
		}
	} else {
		// upload file
		if err = pushFile(ctx, client, whereDataStorage, srcRelPath, srcAbsPath, remoteObjectPrefix); err != nil {
			return err
		}
	}

	return nil
}

func pushDir(ctx context.Context, client dataset.Client, whereDataStorage generated.DataStorageWhereUniqueInput, srcRelPath string, srcAbsPath string, remoteObjectPrefix string) error {

	fileCountChan := make(chan int, 1)
	resultChan := make(chan interface{})
	errChan := make(chan error, 1)
	defer close(fileCountChan)
	defer close(resultChan)

	go func() {
		if err := client.UploadDir(ctx,
			whereDataStorage,
			dataset.UploadDirInput{SrcAbsPath: srcAbsPath, RemoteObjectPrefix: remoteObjectPrefix},
			fileCountChan,
			resultChan,
			nil,
		); err != nil {
			errChan <- err
		}
	}()

	fileCount := <-fileCountChan
	uploadedCount := 0

	bar := progressbar.NewOptions(fileCount,
		progressbar.OptionSetDescription(fmt.Sprintf("%s -> %s", srcRelPath, remoteObjectPrefix)),
		progressbar.OptionFullWidth(),
		progressbar.OptionShowCount(),
	)
	defer fmt.Printf("\n")

	for range resultChan {
		uploadedCount++
		err := bar.Add(1)
		if err != nil {
			return err
		}
		if uploadedCount == fileCount {
			break
		}
		select {
		case err := <-errChan:
			return err
		default: // do nothing
		}
	}

	return nil
}

func pushFile(ctx context.Context, client dataset.Client, whereDataStorage generated.DataStorageWhereUniqueInput, srcRelPath string, srcAbsPath string, remoteObjectKey string) error {

	spinner := spinner_utils.NewAPICallSpinner()
	spinner.Prefix = fmt.Sprintf("Uploading %s -> %s ", srcRelPath, remoteObjectKey)
	spinner.FinalMSG = fmt.Sprintf("Uploaded %s -> %s\n", srcRelPath, remoteObjectKey)

	spinner.Start()
	defer spinner.Stop()

	return client.UploadFile(ctx, whereDataStorage, dataset.UploadFileInput{SrcAbspath: srcAbsPath, RemoteObjectKey: remoteObjectKey})
}
