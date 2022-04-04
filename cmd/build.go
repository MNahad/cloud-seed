package cmd

import (
	build "github.com/mnahad/cloud-seed/build"
	"github.com/spf13/cobra"
)

var env *string

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the application",
	Long:  "Package the code and synthesise the infrastructure",
	Run: func(cmd *cobra.Command, args []string) {
		build.Build(*env)
	},
}

func init() {
	env = buildCmd.Flags().String("env", "", "Set the environment")
	buildCmd.MarkFlagRequired("env")
}
