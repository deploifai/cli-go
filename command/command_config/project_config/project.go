package project_config

import "fmt"

type ProjectNotInitializedError struct{}

func (r ProjectNotInitializedError) Error() string {
	return fmt.Sprintf("project is not initialized, please run 'deploifai project init' to initialize")
}

type Project struct {
	ID string `toml:"id"`
}

func (r *Project) IsInitialized() bool {
	return r.ID != ""
}
