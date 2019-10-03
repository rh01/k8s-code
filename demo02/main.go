package main

import (
	"flag"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path"
)

func main() {
	// 定义一些变量
	var (
		kubeconfig *string
	)

	// 初始化客户端
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", path.Join(home, ".kube", "config"), "(Optional) absoult path to the kubeconfig")

	} else {
		kubeconfig = flag.String("kubeconfig", "", "(Optional) absoult path to the kubeconfig")
	}
	flag.Parse()

	// build rest.config from flags
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// 创建一个新的clientset对象
	clientset, err := kubernetes.NewForConfig(config)
	if err != err{
		panic(err.Error())
	}

	// 通过访问corev1来遍历pod
	podList, err := clientset.CoreV1().Pods("default").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(*podList)

}
