package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete cluster",
	Run: func(cmd *cobra.Command, args []string) {
		if err := deleteCluster(); err != nil {
			log.Fatalf("delete cluster error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func deleteCluster() error {
	cmd := exec.Command("kind", "delete", "cluster")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("delete cmd error: %w", err)
	}
	return nil
}
