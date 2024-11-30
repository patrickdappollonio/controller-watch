package main

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

func getMainFunction() *cobra.Command {
	var kubeconfig string

	cmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(kubeconfig, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}

	cmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "path to the kubeconfig file to use for this controller")

	return cmd
}

func run(kubeconfigPath string, screen, stderr io.Writer) error {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to build configuration: %w", err)
	}

	clientset, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	gitopsRepoCRD := clientset.Resource(
		schema.GroupVersionResource{
			Group:    "kubefirst.konstruct.io",
			Version:  "v1beta1",
			Resource: "GitopsCatalog",
		},
	)
	// list all pods

	watcher, err := gitopsRepoCRD.Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list catalog repos: %w", err)
	}
	defer watcher.Stop()

	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Added:
			pod, ok := event.Object.(*v1.Pod)
			if !ok {
				return fmt.Errorf("unexpected object type: %T", event.Object)
			}

			onAdd(clientset, screen, pod.DeepCopy())
		case watch.Modified:
			pod, ok := event.Object.(*v1.Pod)
			if !ok {
				return fmt.Errorf("unexpected object type: %T", event.Object)
			}

			onModify(clientset, screen, pod.DeepCopy())
		case watch.Deleted:
			pod, ok := event.Object.(*v1.Pod)
			if !ok {
				return fmt.Errorf("unexpected object type: %T", event.Object)
			}

			onDelete(clientset, screen, pod.DeepCopy())
		case watch.Error:
			return fmt.Errorf("watch error: %v", event.Object)
		}
	}

	return nil
}
