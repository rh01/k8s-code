package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	appsv1beta1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	yaml2 "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path"
)

func main() {
	// 定义一些变量
	var (
		kubeconfig *string
		deployJson, deployYaml []byte
		config     *rest.Config
		clientSet  *kubernetes.Clientset
		deployment = appsv1beta1.Deployment{}
		err error
	)

	// 初始化客户端
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", path.Join(home, ".kube", "config"), "(Optional) absoult path to the kubeconfig")

	} else {
		kubeconfig = flag.String("kubeconfig", "", "(Optional) absoult path to the kubeconfig")
	}
	flag.Parse()

	// build rest.config from flags
	if config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig); err != nil {
		panic(err.Error())
	}

	// 读取yaml文件
	if deployYaml, err = ioutil.ReadFile("./nginx.yaml"); err != nil {
		panic(err.Error())
	}

	// YAML转json
	if deployJson, err = yaml2.ToJSON(deployYaml); err != nil {
		panic(err.Error())
	}

	// JSON转struct
	if err := json.Unmarshal(deployJson, &deployment); err != nil {
		panic(err.Error())
	}

	// 创建clientset对象
	if clientSet, err = kubernetes.NewForConfig(config); err != nil {
		panic(err.Error())
	}

	// 修改replicas的数量为1
	var replicas int32 = 1
	deployment.Spec.Replicas = &replicas

	// 查询k8s是否有该deployment
	if _, err := clientSet.AppsV1().Deployments("default").Get(deployment.Name, metav1.GetOptions{}); err != nil {
		// 若存在
		if !errors.IsNotFound(err)  {
			panic("Deployment founded.")
		}

		// 不存在则创建
		if _, err := clientSet.AppsV1().Deployments("default").Create(&deployment); err != nil {
			panic(err.Error())
		}
	} else {
		// 更if _, err := clientSet.AppsV1().Deployments("default").Update(&deployment); err != nil {
			panic(err.Error())
		}

	fmt.Println("apply 成功")
	return
}

