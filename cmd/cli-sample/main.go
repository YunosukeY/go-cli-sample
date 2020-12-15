package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/sync/errgroup"
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
			log.Fatalf("setup cluster error: %v", err)
		}
	case "test":
		if err := testCluster(); err != nil {
			log.Fatalf("test cluster error: %v", err)
		}
	case "delete":
		if err := deleteCluster(); err != nil {
			log.Fatalf("delete cluster error: %v", err)
		}
	default:
		printUsage()
		os.Exit(1)
	}
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

func testCluster() error {
	return nil
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

func printUsage() {
	usage := `Usage:
- go run ./cmd/cli-sample create
- go run ./cmd/cli-sample delete`
	fmt.Println(usage)
}
