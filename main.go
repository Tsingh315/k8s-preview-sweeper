package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const expiresKey = "expires-at"

func decide(annotations map[string]string, now time.Time) (string, string) {
	raw, ok := annotations[expiresKey]
	if !ok {
		return "skip", "no expires-at annotation"
	}
	expiresAt, err := time.Parse(time.RFC3339, raw)

	if err != nil {
		return "skip", fmt.Sprintf("bad expires-at %q", raw)
	}
	
	if now.Before(expiresAt) {
		return "keep", "expires " + raw
	}
	return "delete", "expired " + raw
}

func main() {
	dryRun := flag.Bool("dry-run", true, "print what would be deleted without deleting")
	flag.Parse()

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "home dir:" , err)
		os.Exit(1)
	}

	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "kubeconfig:", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "client:", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	list, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: "preview=true",
	})
	if err!= nil {
		fmt.Fprintln(os.Stderr, "list namespaces: ", err)
	}
	
	now := time.Now()
	for _, ns := range list.Items {
		action, reason:= decide(ns.Annotations, now)
		if action != "delete" {
			fmt.Printf("%-6s %s (%s)\n", action, ns.Name, reason)
			continue
		}
		if *dryRun {
			fmt.Printf("Would Delete %s (%s)\n", ns.Name, reason)
			continue
		}	
		if err := clientset.CoreV1().Namespaces().Delete(ctx, ns.Name, metav1.DeleteOptions{}); err!=nil {
			fmt.Fprintf(os.Stderr, "delete %s failed: %v\n", ns.Name, err)
			continue
		}
		fmt.Printf("Deleted %s (%s)\n", ns.Name, reason)
	}
}
