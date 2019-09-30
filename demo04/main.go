package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	yaml2 "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path"
)

func main() {

	var kubeconfig *string

	// 声明并初始化我们的deployment,container对象
	deployment := appsv1.Deployment{}
	var containers []corev1.Container
	var nginxContainer corev1.Container

	// 第一步：从命令行中创建一个clientset，过程如下
	if home := homedir.HomeDir(); home != ""{
		kubeconfig = flag.String("kubeconfig", path.Join(home, ".kube", "config"), "保存kubeconfig的路径")
	}else {
		kubeconfig = flag.String("kubeconfig", "", "保存kubeconfig的路径")
	}

	// 这里我们可以提供masterURI即集群的地址，或者可以提供kubeconfig的路径名
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err!= nil{
		panic(err.Error())
	}

	// 创建clientset，其中clientset支持多种资源
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil{
		panic(err.Error())
	}

	// 读取yaml文件
	deployYaml, err := ioutil.ReadFile("./nginx.yaml")
	if err != nil {
		panic(err.Error())
	}

	// 将yaml文件转换为json
	deployJson, err := yaml2.ToJSON(deployYaml)
	if err != nil{
		panic(err.Error())
	}

	// 将json再转换为struct
	if err := json.Unmarshal(deployJson, &deployment); err != nil {
		panic(err.Error())
	}

	// 定义的container
	nginxContainer.Name = "nginx"
	nginxContainer.Image = "arm32v6/nginx:1.14-alpine"
	containers = append(containers, nginxContainer)

	// 【注意】修改podTemplate，定义container列表
	deployment.Spec.Template.Spec.Containers = containers

	// 更新deployment
	if _, err := clientset.AppsV1().Deployments("default").Update(&deployment); err != nil {
		panic(err.Error())
	}

	fmt.Println("应用成功")
	return
}
