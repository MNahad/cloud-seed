package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "cloud-seed",
	Aliases: []string{"seed"},
	Short:   "Deploy serverless apps that target multi-cloud with ease",
	Long:    "Cloud Seed is a Terraform multi-cloud configuration generator for serverless apps",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
