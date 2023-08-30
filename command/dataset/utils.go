package dataset

import (
	"fmt"
	"github.com/deploifai/cli-go/command/command_config/project_config"
	"github.com/deploifai/cli-go/utils/spinner_utils"
	"github.com/schollz/progressbar/v3"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func getDataset(projectConfig project_config.Config) (ok bool, dataset project_config.Dataset, dirPath string, err error) {

	cwd, err := os.Getwd()
	if err != nil {
		return false, dataset, dirPath, err
	}

	projectDir := filepath.Dir(projectConfig.ConfigFile)

	for _, d := range projectConfig.Datasets {
		datasetDirPath := filepath.Join(projectDir, filepath.FromSlash(d.LocalDirectory))

		ok, err := isSubDir(datasetDirPath, cwd)
		if err != nil {
			return false, dataset, dirPath, err
		} else if ok {
			return true, d, datasetDirPath, nil
		}
	}

	return false, dataset, dirPath, nil
}

func isSubDir(parent string, child string) (bool, error) {

	up := filepath.Join("..", string(filepath.Separator))

	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false, err
	}

	return !strings.HasPrefix(rel, up) && rel != "..", nil

}

func getAbsPaths(relativePaths []string) ([]string, error) {
	if len(relativePaths) == 0 {
		currentWorkingDirectory, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		return []string{currentWorkingDirectory}, nil
	} else {
		return convertToAbsPaths(relativePaths)
	}
}

func convertToAbsPaths(paths []string) (targets []string, err error) {
	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		targets = append(targets, absPath)
	}

	return targets, nil
}

func getRemoteObjectPrefixes(datasetDirPath string, absPaths []string) (remoteObjectPrefix []string, err error) {
	for _, path := range absPaths {
		relativePath, err := filepath.Rel(datasetDirPath, path)
		if err != nil {
			return nil, err
		}
		remoteObjectPrefix = append(remoteObjectPrefix, relativePath)
	}

	return remoteObjectPrefix, nil
}

func runDir(f func(fileCountChan chan<- int, resultChan chan<- interface{}) error, progressBarDescription string) error {

	fileCountChan := make(chan int, 1)
	resultChan := make(chan interface{})
	errChan := make(chan error, 2)
	defer close(fileCountChan)
	defer close(resultChan)
	defer close(errChan)

	var wg sync.WaitGroup
	wg.Add(1)

	go func(errChan chan<- error) {
		defer wg.Done()
		if err := f(fileCountChan, resultChan); err != nil {
			errChan <- err
		}
	}(errChan)

	go func(fileCountChan <-chan int, resultChan <-chan interface{}, errChan chan<- error) {

		totalCount := <-fileCountChan
		currentCount := 0
		bar := progressbar.NewOptions(totalCount,
			progressbar.OptionSetDescription(progressBarDescription),
			progressbar.OptionFullWidth(),
			progressbar.OptionShowCount(),
		)
		defer fmt.Printf("\n")

		// add 0 to start the progress bar
		if err := bar.Add(0); err != nil {
			errChan <- err
			return
		}

		for range resultChan {
			currentCount++
			if err := bar.Add(1); err != nil {
				errChan <- err
				return
			}
			if currentCount == totalCount {
				break
			}
		}
	}(fileCountChan, resultChan, errChan)

	wg.Wait()

	select {
	case err := <-errChan:
		return err
	default: // do nothing
		return nil
	}

}

func runFile(f func() error, prefixMessage string, finalMessage string) error {

	spinner := spinner_utils.NewAPICallSpinner()
	spinner.Prefix = prefixMessage

	spinner.Start()

	if err := f(); err != nil {
		spinner.Stop()
		return err
	}

	spinner.Stop()
	fmt.Println(finalMessage)

	return nil
}
