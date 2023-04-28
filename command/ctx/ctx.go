package ctx

import (
	"github.com/deploifai/cli-go/api"
	"github.com/deploifai/cli-go/command/command_config"
	"github.com/spf13/cobra"
)

type ContextValue struct {
	Config *command_config.Config
	API    *api.API
}

func NewContextValue(config *command_config.Config, api *api.API) *ContextValue {
	return &ContextValue{
		Config: config,
		API:    api,
	}
}

func GetContextValue(cmd *cobra.Command) *ContextValue {
	// perform type assertion
	return cmd.Root().Context().Value("value").(*ContextValue)
}
