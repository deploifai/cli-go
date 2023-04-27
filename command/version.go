package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number.",
	Long:  `Print the version number of this CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Deploifai CLI v0.0.1")
	},
}
