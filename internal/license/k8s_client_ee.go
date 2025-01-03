//go:build enterprise
// +build enterprise

package license

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const PodNameSpace = "POD_NAMESPACE"

// 不采集硬件信息，获取集群的uid和namespace
// Kubernetes 中，kube-system 是一个默认的命名空间，用于运行 Kubernetes 集群的核心组件和系统服务，拥有唯一的UID
func CollectK8sInfo() (string, error) {
	kubeSysNamespace := "kube-system"
	curPodNamespace := os.Getenv(PodNameSpace)
	namespaceUID, err := GetNamespaceUID(kubeSysNamespace)
	if err != nil {
		return "", fmt.Errorf("failed to get kube-system namespace info.err: %v", err)
	}

	return fmt.Sprintf("%s|%v", curPodNamespace, namespaceUID), nil
}

func GetNamespaceUID(namespaceName string) (string, error) {
	n, err := clientSet.CoreV1().Namespaces().Get(context.TODO(), namespaceName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", n.UID), nil
}

var clientSet *kubernetes.Clientset
var _config *rest.Config

func GetClientset() *kubernetes.Clientset {
	return clientSet
}

func GetConfig() *rest.Config {
	return _config
}

func InitClientSet() error {
	var err error
	_config, err = rest.InClusterConfig()
	if err != nil {
		return err
	}

	clientSet, err = kubernetes.NewForConfig(_config)
	if err != nil {
		return err
	}

	return nil
}

// 是否k8s环境
func Ink8sCluster() bool {
	_, err := rest.InClusterConfig()
	return err == nil
}

func init() {
	err := InitClientSet()
	if err != nil {
		fmt.Printf("[INFO] init k8s client set failed: %v", err)
	}
}
