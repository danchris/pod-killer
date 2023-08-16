package main

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	kubeconfig := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)

	// Initialize kubernetes-client
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Printf("Error building kubeconfig: %v\n", err)
		os.Exit(1)
	}

	// create new client with the given config
	// https://pkg.go.dev/k8s.io/client-go/kubernetes?tab=doc#NewForConfig
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		fmt.Printf("Error building kubernetes clientset: %v\n", err)
		os.Exit(2)
	}

	// use the app's label selector name. Remember this should match with
	// the deployment selector's matchLabels. Replace <APPNAME> with the
	// name of your choice
	options := metav1.ListOptions{
		LabelSelector: "pod-killer=1/2",
	}

	// get the pod list
	// https://pkg.go.dev/k8s.io/client-go@v11.0.0+incompatible/kubernetes/typed/core/v1?tab=doc#PodInterface
	podList, _ := kubeClient.CoreV1().Pods("default").List(options)

	// List() returns a pointer to slice, derefernce it, before iterating
	for _, podInfo := range (*podList).Items {
		fmt.Printf("pods-name=%v\n", podInfo.Name)
		fmt.Printf("pods-status=%v\n", podInfo.Status.Phase)
		fmt.Printf("pods-condition=%v\n", podInfo.Status.Conditions)
	}
}
