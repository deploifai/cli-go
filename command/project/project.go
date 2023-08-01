/*
Copyright © 2023 Sean Chok
*/
package project

import (
	"github.com/spf13/cobra"
)

// Cmd represents the project command
var Cmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects.",
	Long: `List, or create projects in the current workspace.

A project on Deploifai refers to a collection of cloud resources that are managed together for a particular Machine Learning project. 
For example, a project may contain a dataset, a training server, an experiment, and a model deployment.
`,
}

func init() {
	Cmd.AddCommand(listCmd, createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
