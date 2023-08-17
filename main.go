package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	kubeconfig := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)

	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Printf("Error building kubeconfig: %v\n", err)
		os.Exit(1)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		fmt.Printf("Error building kubernetes clientset: %v\n", err)
		os.Exit(2)
	}

	parsedSelector, err := labels.Parse("pod-killer/alive, pod-killer/name")
	if err != nil {
		fmt.Printf("Error bad label: %v", err)
	}

	options := metav1.ListOptions{
		LabelSelector: parsedSelector.String(),
	}

	ticker := time.NewTicker(1000 * time.Millisecond)
	done := make(chan bool)

	go func() {

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				podList, _ := kubeClient.CoreV1().Pods("default").List(context.TODO(), options)

				var toBeKilled []string

				masters := pickMasters(podList)

				for _, podInfo := range (*podList).Items {
					name := podInfo.ObjectMeta.Labels["pod-killer/name"]
					aliveLabel, err := strconv.Atoi(podInfo.ObjectMeta.Labels["pod-killer/alive"])
					if err != nil {
						fmt.Printf("Error to get pod-killer/alive label %v\n", err)
					}
					currentAlive := getCurrentAlive(podList, name)
					toBeKilled = append(toBeKilled, pickCandidates(podList, name, currentAlive-aliveLabel, masters)...)
				}

				for _, v := range toBeKilled {
					err := kubeClient.CoreV1().Pods("default").Delete(context.TODO(), v, metav1.DeleteOptions{})
					if err != nil {
						fmt.Printf("Error to kill pod %v\n", err)
					}
				}
			}
		}
	}()

	sig := <-sigs
	fmt.Printf("signal %v received. Exiting gracefully\n", sig)

}

func pickMasters(podList *v1.PodList) map[string]string {
	masters := make(map[string]string)
	for _, podInfo := range (*podList).Items {
		if _, ok := masters[podInfo.ObjectMeta.Labels["pod-killer/name"]]; !ok {
			masters[podInfo.ObjectMeta.Labels["pod-killer/name"]] = podInfo.Name
		}
	}

	return masters
}

func getCurrentAlive(podList *v1.PodList, name string) int {

	count := 0
	for _, podInfo := range (*podList).Items {
		if podInfo.ObjectMeta.Labels["pod-killer/name"] == name {
			count++
		}
	}

	return count
}

func pickCandidates(podList *v1.PodList, name string, haveToBeKilled int, masters map[string]string) []string {

	if haveToBeKilled <= 0 {
		return nil
	}

	candidatesNumber := 0
	var candidates []string

	for _, podInfo := range (*podList).Items {
		nameLabel := podInfo.ObjectMeta.Labels["pod-killer/name"]
		if nameLabel == name && podInfo.Name != masters[nameLabel] {
			candidates = append(candidates, podInfo.Name)
			candidatesNumber++
		}
		if haveToBeKilled == candidatesNumber {
			break
		}
	}
	return candidates
}
