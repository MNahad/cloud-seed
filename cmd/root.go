package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "Cloud Seed",
	Short: "Cloud Seed is a Terraform multi-cloud configuration generator for serverless apps",
	Long:  "Deploy serverless apps without worrying about GCP or AWS infrastructure management",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello world")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
