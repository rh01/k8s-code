package demo01

import (
	"bufio"
	"flag"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
	"os"
	"path/filepath"
)

func main() {
	var kubeconfig *string

	// 解析命令行参数， 主要是解析出kubeconfig的路径，默认情况下在 [$HOME/.kube/config]
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"),
			"(Optional) absoult path to the kubeconfig")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// 创建config对象
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// clientset对象，通过clientSet能够找到kubernetes所有原生资源对应的client
	// 获取的一般方式指定group然后制定特定的version，然后根据resource名字来获取到对应的client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// 获取deploymentclient
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault) // default namespace
	// 定义部署对象
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo01-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo01",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo01",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "arm32v6/nginx:1.14-alpine",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	// 创建部署
	fmt.Println("creating deployment....")
	result, err := deploymentsClient.Create(deployment)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Created deployment %q\n", result.GetObjectMeta().GetName())

	// 更新部署
	prompt()
	fmt.Println("Updating deployment...")
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, getErr := deploymentsClient.Get("demo01-deployment", metav1.GetOptions{})
		if getErr != nil{
			panic(fmt.Errorf("failed to get lastest version of Deployment: %v", getErr))
		}
		result.Spec.Replicas = int32Ptr(1)
		// 减少副本数量
		result.Spec.Template.Spec.Containers[0].Image = "arm32v6/nginx:1.17-alpine" // 更换最新的版本
		_, updateErr := deploymentsClient.Update(result)
		if updateErr != nil {
			panic(updateErr.Error())
		}
		return updateErr
	})
	if retryErr != nil {
		panic(err)
	}

	// 列出部署
	prompt()
	fmt.Printf("Listing deployments in namespace %q\n", apiv1.NamespaceDefault)
	list, err := deploymentsClient.List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, d := range list.Items {
		fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
	}

	// 删除部署
	fmt.Println("Deleting deployment...")
	deletionPolicy := metav1.DeletePropagationForeground
	if err := deploymentsClient.Delete("demo01-deployment", &metav1.DeleteOptions{
		PropagationPolicy: &deletionPolicy,
	}); err != nil {
		panic(err)
	}
	fmt.Println("Deleted deployment.")
}

func prompt() {
	fmt.Printf("-> Press Return key to continue")
	scanner := bufio.NewScanner(os.Stdin)

	// 如果没有数据就一直循环，直到有数据为止
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err.Error())
	}

	fmt.Println()
}

func int32Ptr(i int32) *int32 { return &i }
