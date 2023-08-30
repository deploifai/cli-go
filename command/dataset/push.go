/*
Copyright © 2023 Sean Chok
*/
package dataset

import (
	"context"
	"errors"
	"fmt"
	"github.com/deploifai/cli-go/command/ctx"
	"github.com/deploifai/sdk-go/api/generated"
	"github.com/deploifai/sdk-go/service/dataset"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push [<path>...]",
	Short: "Push local files to a dataset",
	Long: `Upload local files to a dataset.

This requires the local directory to be initialised as a dataset first.
Use the command "deploifai dataset init" to do that.

Each <path> can be a directory or a file.
If no <path> is specified, the current directory is used.
`,
	RunE: func(cmd *cobra.Command, args []string) error {

		_context := ctx.GetContextValue(cmd)

		// get the dataset and directory path from config
		ok, ds, datasetDirPath, err := getDataset(*_context.Project)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("the current directory is not initialised as a dataset")
		}

		// get the source absolute paths from args
		srcAbsPaths, err := getAbsPaths(args)
		if err != nil {
			return err
		}

		if ok, invalidArgs, err := verifyPushPaths(datasetDirPath, args, srcAbsPaths); err != nil {
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

func verifyPushPaths(datasetDirPath string, args []string, paths []string) (ok bool, invalidArgs []string, err error) {
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

	return len(invalidArgs) == 0, invalidArgs, nil
}

func push(ctx context.Context, client dataset.Client, dataStorageId string, srcRelPath string, srcAbsPath string, remoteObjectPrefix string) error {

	fileInfo, err := os.Stat(srcAbsPath)
	if err != nil {
		return err
	}

	whereDataStorage := generated.DataStorageWhereUniqueInput{ID: &dataStorageId}

	if fileInfo.IsDir() {
		// upload directory
		return pushDir(ctx, client, whereDataStorage, srcRelPath, srcAbsPath, remoteObjectPrefix)
	} else {
		// upload file
		return pushFile(ctx, client, whereDataStorage, srcRelPath, srcAbsPath, remoteObjectPrefix)
	}
}

func pushDir(ctx context.Context, client dataset.Client, whereDataStorage generated.DataStorageWhereUniqueInput, srcRelPath string, srcAbsPath string, remoteObjectPrefix string) error {

	f := func(fileCountChan chan<- int, resultChan chan<- interface{}) error {
		return client.UploadDir(ctx,
			whereDataStorage,
			dataset.UploadDirInput{SrcAbsPath: srcAbsPath, RemoteObjectPrefix: remoteObjectPrefix},
			fileCountChan,
			resultChan,
			nil,
		)
	}

	return runDir(f, fmt.Sprintf("%s -> %s", srcRelPath, remoteObjectPrefix))

}

func pushFile(ctx context.Context, client dataset.Client, whereDataStorage generated.DataStorageWhereUniqueInput, srcRelPath string, srcAbsPath string, remoteObjectKey string) error {

	f := func() error {
		return client.UploadFile(ctx, whereDataStorage, dataset.UploadFileInput{SrcAbspath: srcAbsPath, RemoteObjectKey: remoteObjectKey})
	}

	prefixMessage := fmt.Sprintf("Uploading %s -> %s ", srcRelPath, remoteObjectKey)
	finalMessage := fmt.Sprintf("Uploaded %s -> %s", srcRelPath, remoteObjectKey)

	return runFile(f, prefixMessage, finalMessage)

}
