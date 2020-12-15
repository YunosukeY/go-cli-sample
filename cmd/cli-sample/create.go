package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create cluster and deploy resources",
	Run: func(cmd *cobra.Command, args []string) {
		if err := setupCluster(); err != nil {
			log.Fatalf("setup cluster error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}

func setupCluster() error {
	if err := createCluster(); err != nil {
		return fmt.Errorf("create cluster error: %w", err)
	}

	if err := deployResources(); err != nil {
		return fmt.Errorf("deploy resources error: %w", err)
	}

	return nil
}

func createCluster() error {
	cmd := exec.Command("kind", "create", "cluster")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("create cmd error: %w", err)
	}
	return nil
}

type resource struct {
	kind string
	name string
}

func deployResources() error {
	repodir, err := repoDir()
	if err != nil {
		return fmt.Errorf("repo dir error: %w", err)
	}

	if err = apply(filepath.Join(repodir, "k8s", "namespace.yaml")); err != nil {
		return fmt.Errorf("apply namespace error: %w", err)
	}
	if err = apply(filepath.Join(repodir, "k8s", "deployment.yaml")); err != nil {
		return fmt.Errorf("apply deployment error: %w", err)
	}

	rscs := []resource{
		{
			kind: "deployment",
			name: "mysql",
		},
		{
			kind: "deployment",
			name: "redis",
		},
	}
	if err = wait(rscs, "test-ns", "available", "10m"); err != nil {
		return err
	}

	return nil
}

func repoDir() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("git cmd error: %w", err)
	}
	out = out[:len(out)-1] // 末尾の改行を削除
	return string(out), nil
}

func apply(manifest string) error {
	cmd := exec.Command("kubectl", "apply", "-f", manifest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("apply cmd error: %w", err)
	}
	return nil
}

// goroutineを使わなくても良いが，どのリソースのwaitが失敗したかを調べる必要がある
func wait(rscs []resource, namespace string, cond string, timeout string) error {
	eg := errgroup.Group{}
	for _, rsc := range rscs {
		rsc := rsc // https://golang.org/doc/faq#closures_and_goroutines
		eg.Go(func() error {
			if err := _wait(rsc, namespace, cond, timeout); err != nil {
				// waitが失敗したリソースをdescribeで見る
				if err = describe(rsc, namespace); err != nil {
					return fmt.Errorf("describe error: %w", err)
				}
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func _wait(rsc resource, namespace string, cond string, timeout string) error {
	out, err := exec.Command("kubectl", "wait", rsc.kind+"/"+rsc.name, "-n", namespace, "--for=condition="+cond, "--timeout="+timeout).CombinedOutput()
	fmt.Print(string(out))
	if err != nil {
		return fmt.Errorf("wait cmd error: %w", err)
	}
	return nil
}

// TODO: Deployment以外のリソース
func describe(rsc resource, namespace string) error {
	// describe deployment
	out, err := exec.Command("kubectl", "describe", rsc.kind+"/"+rsc.name, "-n", namespace).Output()
	fmt.Print(string(out))
	if err != nil {
		return fmt.Errorf("describe deployment error: %w", err)
	}

	// describe pod
	podname, err := exec.Command("sh", "-c", "kubectl get pod -A -o name | grep "+rsc.name).Output()
	if err != nil {
		return fmt.Errorf("get pod error: %w", err)
	}
	podname = podname[:len(podname)-1] // 末尾の改行を削除
	// TODO: podが2以上ある時
	out, err = exec.Command("kubectl", "describe", string(podname), "-n", namespace).Output()
	fmt.Print(string(out))
	if err != nil {
		return fmt.Errorf("describe pod error: %w", err)
	}

	return nil
}
