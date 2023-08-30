package ctx

import (
	"github.com/deploifai/cli-go/command/command_config/project_config"
	"github.com/deploifai/cli-go/command/command_config/root_config"
	"github.com/deploifai/sdk-go/config"
	"github.com/spf13/cobra"
)

type ContextValue struct {
	Root                *root_config.Config
	Project             *project_config.Config
	ServiceClientConfig *config.Config
}

func NewContextValue(root *root_config.Config, projectConfig *project_config.Config, serviceClientConfig *config.Config) *ContextValue {
	return &ContextValue{
		Root:                root,
		Project:             projectConfig,
		ServiceClientConfig: serviceClientConfig,
	}
}

func GetContextValue(cmd *cobra.Command) *ContextValue {
	// perform type assertion
	return cmd.Root().Context().Value("value").(*ContextValue)
}
