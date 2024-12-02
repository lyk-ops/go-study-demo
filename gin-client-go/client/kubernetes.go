package client

import (
	"flag"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

func GetK8sClientSet() (*kubernetes.Clientset, error) {
	config, err := GetRestConfig()
	if err != nil {
		klog.Fatal(err)
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
		return nil, err
	}
	return clientSet, nil
}
func GetRestConfig() (config *rest.Config, err error) {
	var kubeConfig *string
	if home := homedir.HomeDir(); home != "" {
		fmt.Printf("home:%v", home)
		kubeConfig = flag.String("kubeconfig", ".kube/config", "absolute path to the kubeconfig file")
	} else {
		klog.Fatal("Config not found")
		return
	}
	flag.Parse()
	config, err = clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		klog.Fatal(err)
		return
	}
	return
}
