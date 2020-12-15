package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		printUsage()
		os.Exit(1)
	}

	switch args := flag.Args()[0]; args {
	case "create":
		if err := setupCluster(); err != nil {
			log.Fatal(err)
		}
	case "test":
		if err := testCluster(); err != nil {
			log.Fatal(err)
		}
	case "delete":
		if err := deleteCluster(); err != nil {
			log.Fatal(err)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func setupCluster() error {
	return runCmd("kind", "create", "cluster")
}

func testCluster() error {
	return nil
}

func deleteCluster() error {
	return runCmd("kind", "delete", "cluster")
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cmd run error: %w", err)
	}
	return nil
}

func printUsage() {
	usage := `Usage:
- kind.sh create
- kind.sh delete`
	fmt.Println(usage)
}
