package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

func onAdd(_ kubernetes.Interface, screen io.Writer, pod *v1.Pod) {
	fmt.Fprintf(screen, "POD ADDED: name: %q namespace: %q\n", pod.Name, pod.Namespace)
}

func onModify(client kubernetes.Interface, screen io.Writer, pod *v1.Pod) {
	fmt.Fprintf(screen, "POD MODIFIED: name: %q namespace: %q\n", pod.Name, pod.Namespace)

	t := pod.ObjectMeta.GetDeletionTimestamp()
	if t == nil || t.IsZero() {
		fmt.Fprintf(screen, "Pod is not candidate for deletion, not touching it\n")
		return
	}

	fmt.Fprintf(screen, "Pod is candidate for deletion, patching finalizers: %s\n", strings.Join(pod.Finalizers, ", "))

	newFinalizers := make([]string, 0, len(pod.Finalizers))
	found := false
	for _, finalizer := range pod.Finalizers {
		if finalizer == "konstruct.kubefirst.io/muse-mulatu" {
			found = true
			continue
		}
		newFinalizers = append(newFinalizers, finalizer)
	}

	if !found {
		fmt.Fprintf(screen, "Finalizer not found, nothing to do\n")
		return
	}

	patch := map[string]interface{}{
		"metadata": map[string]interface{}{
			"finalizers": newFinalizers,
		},
	}
	patchJson, err := json.Marshal(patch)
	if err != nil {
		fmt.Fprintf(screen, "Unable to marshal string: %s", err)
	}
	updated, err := client.CoreV1().Pods(pod.Namespace).Patch(context.TODO(), pod.Name, types.MergePatchType, patchJson, metav1.PatchOptions{})
	if err != nil {
		fmt.Fprintf(screen, "Failed to update pod: %v\n", err)
		return
	}

	fmt.Fprintf(screen, "Finalizer removed. Current finalizers are: %s\n", strings.Join(updated.Finalizers, ", "))
}

func onDelete(_ kubernetes.Interface, screen io.Writer, pod *v1.Pod) {
	fmt.Fprintf(screen, "POD DELETED: name: %q namespace: %q\n", pod.Name, pod.Namespace)
}
