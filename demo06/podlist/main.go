package main

import (
	"flag"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"

	"os"
	"path/filepath"
)

// This program lists the pods in a cluster equivalent to
//
// kubectl get pods
//
func main() {
	var ns string
	flag.StringVar(&ns, "namespace", "default", "specific namespace")

	flag.Parse()

	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	log.Println("Using kubeconfig file: ", kubeconfig)
	log.Printf("Use namespace %s", ns)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err!=nil{
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err!=nil{
		log.Fatal(err)
	}
	podList, err := clientset.CoreV1().Pods(ns).List(metav1.ListOptions{})
	if err != nil{
		log.Fatal(err)
	}

	for i, item := range podList.Items {
		fmt.Printf("%d - %s\n", i, item.Name)
	}

}