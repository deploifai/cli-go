package ctx

import (
	"github.com/deploifai/cli-go/command/command_config"
	"github.com/deploifai/sdk-go/config"
	"github.com/spf13/cobra"
)

type ContextValue struct {
	Config              *command_config.Config
	ServiceClientConfig *config.Config
}

func NewContextValue(config *command_config.Config, serviceClientConfig *config.Config) *ContextValue {
	return &ContextValue{
		Config:              config,
		ServiceClientConfig: serviceClientConfig,
	}
}

func GetContextValue(cmd *cobra.Command) *ContextValue {
	// perform type assertion
	return cmd.Root().Context().Value("value").(*ContextValue)
}
