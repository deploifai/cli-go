/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cloud_profile

import (
	"github.com/deploifai/cli-go/command/cloud_profile/create"
	"github.com/spf13/cobra"
)

// Cmd represents the cloud-profile command
var Cmd = &cobra.Command{
	Use:     "cloud-profile",
	Aliases: []string{"cp"},
	Short:   "Mange cloud profiles of the current workspace.",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

func init() {
	Cmd.AddCommand(create.Cmd, listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cloudProfileCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cloudProfileCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
