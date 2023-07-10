package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

const VERSION = "v0.5.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number.",
	Long:  `Print the version number of this CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Deploifai CLI %s\n", VERSION)
	},
}
