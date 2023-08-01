/*
Copyright © 2023 Sean Chok
*/
package project

import (
	"context"
	"github.com/deploifai/sdk-go/service/project"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project in the current workspace.",
	Long: `Create a new project on Deploifai for a new Machine Learning project.

Every project requires a cloud profile to first create a project-specific dataset that can be used to store the results of experiments.
`,
	Args: cobra.ExactArgs(1), // requires exactly 1 arg, which is the name of the new cloud profile
	RunE: func(cmd *cobra.Command, args []string) error {

		projectName := args[0]
		cmd.Println("Creating project", projectName)

		return nil
	},
}

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func checkCollision(c context.Context, client project.Client, ) {

}