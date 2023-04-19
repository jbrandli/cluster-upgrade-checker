package main

import (
	"context"
	"log"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Use in-cluster configuration if running inside a cluster
	config, err := rest.InClusterConfig()
	if err != nil {
		homeDir := os.Getenv("HOME")
		// Use the current context in kubeconfig if running outside a cluster
		kubeConfigPath := homeDir + "/.kube/config"
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			log.Fatalf("Failed to get kubeconfig: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to list namespaces: %v", err)
	}

	haDeployments := []string{}
	noPdbDeployments := []string{}
	misconfiguredDeployments := []string{}
	goodDeployments := []string{}
	for _, namespace := range namespaces.Items {
		deployments, err := clientset.AppsV1().Deployments(namespace.Name).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			log.Fatalf("Failed to list deployments in namespace %s: %v", namespace.Name, err)
		}

		for _, deployment := range deployments.Items {
			pdbName := deployment.Name
			pdb, err := clientset.PolicyV1().PodDisruptionBudgets(namespace.Name).Get(context.Background(), pdbName, metav1.GetOptions{})
			if err != nil {
				noPdbDeployments = append(noPdbDeployments, deployment.Name)
				continue
			}
			minAvailable := 0
			maxUnavailable := 0
			if pdb.Spec.MinAvailable == nil {
				if pdb.Spec.MaxUnavailable == nil {
					misconfiguredDeployments = append(misconfiguredDeployments, deployment.Name)
					continue
				} else {
					maxUnavailable = pdb.Spec.MaxUnavailable.IntValue()

				}
			}

			replicas := int(*deployment.Spec.Replicas)
			if minAvailable <= 1 && maxUnavailable <= 1 {

				if replicas >= 2 {
					haDeployments = append(haDeployments, deployment.Name)
				} else {
					goodDeployments = append(goodDeployments, deployment.Name)
				}
			} else {
				misconfiguredDeployments = append(misconfiguredDeployments, deployment.Name)
			}
		}
	}
	writeToFile("good_deployments.txt", goodDeployments)
	writeToFile("ha_deployments.txt", haDeployments)
	writeToFile("no_pdb_deployments.txt", noPdbDeployments)
	writeToFile("misconfigured_deployments.txt", misconfiguredDeployments)
}

// Function write lists of deployments to a file
func writeToFile(filename string, data []string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	for _, deployment := range data {
		f.WriteString(deployment + "\n")
	}
}
