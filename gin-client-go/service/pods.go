package service

import (
	"context"
	"gin-client-go/client"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func GetPods() ([]v1.Pod, error) {
	clientSet, err := client.GetK8sClientSet()
	if err != nil {
		klog.Fatal(err)
		return nil, err
	}
	list, err := clientSet.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Fatal(err)
	}
	return list.Items, nil

}
#add
