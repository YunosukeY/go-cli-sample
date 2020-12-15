package main

import (
	"log"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test cluster",
	Run: func(cmd *cobra.Command, args []string) {
		if err := testCluster(); err != nil {
			log.Fatalf("test cluster error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}

func testCluster() error {
	return nil
}
