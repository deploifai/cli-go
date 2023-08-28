/*
Copyright © 2023 Sean Chok
*/
package dataset

import (
	"github.com/spf13/cobra"
)

// Cmd represents the dataset command
var Cmd = &cobra.Command{
	Use:   "dataset",
	Short: "Manage, and interact with datasets",
	Long: `Initialize, push, or pull datasets in the current workspace.

A dataset refers to a collection of files that are stored in a remote object storage on the cloud.
`,
}

func init() {
	Cmd.AddCommand(initCmd, pushCmd, pullCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
