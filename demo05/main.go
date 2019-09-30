package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	yaml2 "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path"
	"strconv"
	"time"
)

func main() {
	var kubeconfig *string

	// 声明并初始化我们的deployment,container对象
	deployment := appsv1.Deployment{}

	// 第一步：从命令行中创建一个clientset，过程如下
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", path.Join(home, ".kube", "config"), "保存kubeconfig的路径")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "保存kubeconfig的路径")
	}

	// 这里我们可以提供masterURI即集群的地址，或者可以提供kubeconfig的路径名
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// 创建clientset，其中clientset支持多种资源
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// 读取yaml文件
	deployYaml, err := ioutil.ReadFile("./nginx.yaml")
	if err != nil {
		panic(err.Error())
	}

	// 将yaml文件转换为json
	deployJson, err := yaml2.ToJSON(deployYaml)
	if err != nil {
		panic(err.Error())
	}

	// 将json再转换为struct
	if err := json.Unmarshal(deployJson, &deployment); err != nil {
		panic(err.Error())
	}

	// 查询是否存在不存在则创建
	if _, err := clientset.AppsV1().Deployments("default").Create(&deployment); err != nil {
		// 不存在则创建
		if _, err := clientset.AppsV1().Deployments("default").Create(&deployment); err != nil {
			panic(err.Error())
		}
	}

	// 给pod添加标签label
	deployment.Spec.Template.Labels["deploy_time"] = strconv.Itoa(int(time.Now().Unix()))


	// 更新deployment
	if _, err := clientset.AppsV1().Deployments("default").Update(&deployment); err != nil {
		panic(err.Error())
	}

	// 等待更新完成
	for {

		// 获取k8s中的deployment的状态
		k8sDeployment, err := clientset.AppsV1().Deployments("default").Get(deployment.Name, metav1.GetOptions{})
		if err != nil {
			time.Sleep(1 * time.Second)
		}

		// 进行状态判定
		// 1. 如果当前状态的UpdateReplics与我们yaml中定义的一样
		// 2. 如果当前deployment的副本数与yaml中定义的一样
		// 3. 如果可用的deployment的数目与yaml中定义的一样
		// 4. 如果observedGeneration与Generation一样
		// 满足上面四种即可表明我们的应用更新完成
		if k8sDeployment.Status.UpdatedReplicas == *(k8sDeployment.Spec.Replicas) &&
			k8sDeployment.Status.Replicas == *(k8sDeployment.Spec.Replicas) &&
			k8sDeployment.Status.AvailableReplicas == *(k8sDeployment.Spec.Replicas) &&
			k8sDeployment.Generation == k8sDeployment.Generation {
			// 升级完成
			break
		}

		fmt.Printf("部署中：（%d/%d）\n", k8sDeployment.Status.AvailableReplicas, *(k8sDeployment.Spec.Replicas))
		time.Sleep(1*time.Second)
	}

	// 部署成功
	fmt.Println("部署成功！")

	// 打印每一个POd的状态（可能会打印termination中的POd，但最终只会展示新的Pod的列表）
	if podList, err := clientset.CoreV1().Pods("default").List(metav1.ListOptions{
		LabelSelector: "app=nginx",
	}); err == nil {
		for _, pod := range podList.Items {
			podName := pod.Name
			podStatus := string(pod.Status.Phase)

			// 如果当前的pod的状态是podRUnning意味着这个pod已经绑定到一个node上面了，并且该pod的所有容器已经运行了
			// 至少一个容器正在运行或者容器的进程刚刚开始
			if podStatus == string(corev1.PodRunning) {
				 // 汇总错误原因不为空
				if pod.Status.Reason != "" {
					podStatus = pod.Status.Reason
					fmt.Printf("[name:%s status:%s]\n", podName, podStatus)
				}

				// condition有错误信息
				for _, cond := range pod.Status.Conditions {
					// 如果当前的pod处于就绪状态
					if cond.Type == corev1.PodReady {
						// 当前status不为true
						if cond.Status != corev1.ConditionTrue {
							podStatus = cond.Reason
						}
						fmt.Printf("[name:%s status:%s]\n", podName, podStatus)
						break
					}
				}
				// 没有ready condtion，状态未知
				podStatus = "Unknow"
			}
		}
	}
	return
}
