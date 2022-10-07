package cmd

import (
	"github.com/mnahad/cloud-seed/build"
	"github.com/spf13/cobra"
)

var env string
var dir string

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the IaC",
	Long:  "Package the code and synthesise the infrastructure",
	Run: func(cmd *cobra.Command, args []string) {
		build.Build(env, dir)
	},
}

func init() {
	flags := buildCmd.Flags()
	flags.StringVarP(&env, "environment", "e", "", "Set an environment")
	flags.StringVarP(&dir, "project-dir", "d", "", "Select a project directory")
}
