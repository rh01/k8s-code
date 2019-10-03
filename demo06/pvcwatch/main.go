package main

import (
	"flag"
	"fmt"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var ns, label, maxClaims, field string

	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	log.Printf("Use kubeconfig: %s\n", kubeconfig)

	flag.StringVar(&ns, "namespace", "", "show namespace")
	flag.StringVar(&kubeconfig, "kubeconfig", kubeconfig, `kubeconfig file (default "/Users/<username>/.kube/config")`)
	flag.StringVar(&label, "label", "", "Label Selector")
	flag.StringVar(&maxClaims, "max-claims", "20Gi", `Maximum total claims to watch (default "200Gi")`)
	flag.StringVar(&field, "field", "", "Field selector")
	flag.Parse()

	log.Printf("use namespace: %s\n", ns)
	log.Printf("use kubeconfig: %s\n", kubeconfig)
	log.Printf("use label: %s\n", label)
	log.Printf("use max-claims: %s\n", maxClaims)
	log.Printf("use field: %s\n", field)

	//total resource quantities
	var totalClaimedQuant resource.Quantity
	maxClaimedQuant := resource.MustParse(maxClaims)

	// build connection with external cluster
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// initial list
	listOptions := metav1.ListOptions{LabelSelector: label, FieldSelector: field}
	pvcs, err := clientset.CoreV1().PersistentVolumeClaims(ns).List(listOptions)
	if err != nil {
		log.Fatal(err)
	}

	printPVCs(pvcs)
	fmt.Println()

	api := clientset.CoreV1()
	listOptions = metav1.ListOptions{LabelSelector: label,
		FieldSelector: field}
	watcher, err := api.PersistentVolumeClaims(ns).Watch(listOptions)
	if err != nil {
		log.Fatal(err)
	}
	// watch ch
	ch := watcher.ResultChan()

	// Loop through events
	for event := range ch {
		pvc, ok := event.Object.(*v1.PersistentVolumeClaim)
		if !ok {
			log.Fatal("unexpected type")
		}
		quant := pvc.Spec.Resources.Requests[v1.ResourceStorage]
		switch event.Type {
		case watch.Added:
			totalClaimedQuant.Add(quant)
			log.Printf("PVC %s added, claim size %s\n",
				pvc.Name, quant.String())
			if totalClaimedQuant.Cmp(maxClaimedQuant) == 1 {
				log.Printf(
					"\nClaim overage reached: max %s at %s",
					maxClaimedQuant.String(),
					totalClaimedQuant.String())
				// trigger action
				log.Println("*** Taking action ***")
			}

		case watch.Deleted:
			totalClaimedQuant.Sub(quant)
			log.Printf("PVC %s removed, size %s\n",
				pvc.Name, quant.String())
			if totalClaimedQuant.Cmp(maxClaimedQuant) <= 0 {
				log.Printf("Claim usage normal: max %s at %s",
					maxClaimedQuant.String(),
					totalClaimedQuant.String(),
				)
				// trigger action
				log.Println("*** Taking action ***")
			}

		}
	}

}

// printPVCs prints a list of PersistentVolumeClaim on console
func printPVCs(pvcs *v1.PersistentVolumeClaimList) {
	if len(pvcs.Items) == 0 {
		log.Println("No claims found")
		return
	}
	template := "%-32s%-8s%-8s\n"
	fmt.Println("--- PVCs ----")
	fmt.Printf(template, "NAME", "STATUS", "CAPACITY")
	var cap resource.Quantity
	for _, pvc := range pvcs.Items {
		quant := pvc.Spec.Resources.Requests[v1.ResourceStorage]
		cap.Add(quant)
		fmt.Printf(template, pvc.Name, string(pvc.Status.Phase), quant.String())
	}

	fmt.Println("-----------------------------")
	fmt.Printf("Total capacity claimed: %s\n", cap.String())
	fmt.Println("-----------------------------")
}
