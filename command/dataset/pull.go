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
	"path/filepath"
	"strings"
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull [<path>...]",
	Short: "Pull files from a dataset to a local filesystem",
	Long: `Download files from a dataset to a local filesystem.

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

		// get the destination absolute paths from args
		destAbsPaths, err := getAbsPaths(args)
		if err != nil {
			return err
		}

		// verify the paths
		if ok, invalidArgs, err := verifyPullPaths(datasetDirPath, args, destAbsPaths); err != nil {
			return err
		} else if !ok {
			return errors.New(fmt.Sprintf("invalid paths: %s", strings.Join(invalidArgs, ", ")))
		}

		// get the remoteObjectPrefixes from destAbsPaths
		remoteObjectPrefixes, err := getRemoteObjectPrefixes(datasetDirPath, destAbsPaths)
		if err != nil {
			return err
		}

		client := dataset.NewFromConfig(*_context.ServiceClientConfig)

		ok, objectTypes, invalid, err := verifyRemoteObjectPrefixes(cmd.Context(), *client, ds.ID, args, remoteObjectPrefixes)
		if err != nil {
			return err
		} else if !ok {
			return errors.New(fmt.Sprintf("no objects found in paths: %s", strings.Join(invalid, ", ")))
		}

		for i, path := range destAbsPaths {
			destRelPath := "."
			objectType := ObjectTypeDirectory
			if len(args) > 0 {
				destRelPath = args[i]
				objectType = objectTypes[i]
			}
			if err = pull(cmd.Context(), *client, ds.ID, objectType, destRelPath, path, remoteObjectPrefixes[i]); err != nil {
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
	// pullCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pullCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func verifyPullPaths(datasetDirPath string, args []string, paths []string) (ok bool, invalidArgs []string, err error) {
	for i, path := range paths {
		// check if path is a subdirectory of datasetDirPath
		if ok, err := isSubDir(datasetDirPath, path); err != nil {
			return false, nil, err
		} else if !ok {
			invalidArgs = append(invalidArgs, args[i])
		}
	}

	return len(invalidArgs) == 0, invalidArgs, nil
}

type objectType uint8

const (
	ObjectTypeDirectory objectType = 1
	ObjectTypeFile      objectType = 2
)

func verifyRemoteObjectPrefixes(ctx context.Context, client dataset.Client, dataStorageId string, args []string, remoteObjectPrefixes []string) (ok bool, objectTypes []objectType, invalid []string, err error) {

	maxResultsPerPage := 1

	for i, remoteObjectPrefix := range remoteObjectPrefixes {
		prefix := dataset.CleanRemoteObjectPrefix(remoteObjectPrefix)

		// special case: prefix is empty, referring to the root, which is always valid
		if prefix == "" {
			objectTypes = append(objectTypes, ObjectTypeDirectory)
			continue
		}

		directoryPager, err := client.NewListObjectsPager(
			ctx,
			generated.DataStorageWhereUniqueInput{ID: &dataStorageId},
			&dataset.ListObjectsInput{Prefix: &prefix, MaxResultsPerPage: &maxResultsPerPage},
		)
		if err != nil {
			return false, nil, nil, err
		}

		if response, err := directoryPager.NextPage(nil); err != nil {
			return false, nil, nil, err
		} else if len(response.Objects) == 0 {

			// check if it is a file
			filePager, err := client.NewListObjectsPager(
				ctx,
				generated.DataStorageWhereUniqueInput{ID: &dataStorageId},
				&dataset.ListObjectsInput{Prefix: &remoteObjectPrefix, MaxResultsPerPage: &maxResultsPerPage},
			)
			if err != nil {
				return false, nil, nil, err
			}

			if response, err := filePager.NextPage(nil); err != nil {
				return false, nil, nil, err
			} else if len(response.Objects) == 1 && response.Objects[0].Key == remoteObjectPrefix {
				objectTypes = append(objectTypes, ObjectTypeFile)
			} else {
				invalid = append(invalid, args[i])
			}
		} else {
			objectTypes = append(objectTypes, ObjectTypeDirectory)
		}

	}

	return len(invalid) == 0, objectTypes, invalid, nil
}

func pull(ctx context.Context, client dataset.Client, dataStorageId string, objectType objectType, destRelPath string, destAbsPath string, remoteObjectPrefix string) error {

	whereDataStorage := generated.DataStorageWhereUniqueInput{ID: &dataStorageId}

	if objectType == ObjectTypeDirectory {
		return pullDir(ctx, client, whereDataStorage, destRelPath, destAbsPath, remoteObjectPrefix)
	} else if objectType == ObjectTypeFile {
		return pullFile(ctx, client, whereDataStorage, destRelPath, destAbsPath, remoteObjectPrefix)
	}

	return nil
}

func pullDir(ctx context.Context, client dataset.Client, whereDataStorage generated.DataStorageWhereUniqueInput, destRelPath string, destAbsPath string, remoteObjectPrefix string) error {

	f := func(fileCountChan chan<- int, resultChan chan<- interface{}) error {
		return client.DownloadDir(ctx,
			whereDataStorage,
			dataset.DownloadDirInput{DestAbsPath: destAbsPath, RemoteObjectPrefix: remoteObjectPrefix},
			fileCountChan,
			resultChan,
			nil,
		)
	}

	return runDir(f, fmt.Sprintf("%s -> %s", remoteObjectPrefix, destRelPath))
}

func pullFile(ctx context.Context, client dataset.Client, whereDataStorage generated.DataStorageWhereUniqueInput, destRelPath string, destAbsPath string, remoteObjectKey string) error {

	f := func() error {
		if err := os.MkdirAll(filepath.Dir(destAbsPath), 0755); err != nil {
			return err
		}

		return client.DownloadFile(ctx, whereDataStorage, dataset.DownloadFileInput{DestAbsPath: destAbsPath, RemoteObjectKey: remoteObjectKey})
	}

	prefixMessage := fmt.Sprintf("Downloading %s -> %s ", remoteObjectKey, destRelPath)
	finalMessage := fmt.Sprintf("Downloaded %s -> %s", remoteObjectKey, destRelPath)

	return runFile(f, prefixMessage, finalMessage)
}
