package main

import (
	"fmt"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"time"
)

func main() {
	config, err := config.GetConfig()
	if err != nil {
		panic(err)
		return
	}
	// Create a new clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
		return
	}
	// Create a new informer factory
	informerFactory := informers.NewSharedInformerFactory(clientSet, 30*time.Second)
	deployInformer := informerFactory.Apps().V1().Deployments()
	informer := deployInformer.Informer()
	deployLister := deployInformer.Lister()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		UpdateFunc: onUpdate,
		DeleteFunc: onDelete,
	})
	stopper := make(chan struct{})
	defer close(stopper)
	// Start the informer List and watch
	informerFactory.Start(stopper)
	// Wait for the initial synchronization of the local cache
	informerFactory.WaitForCacheSync(stopper)
	deployment, err := deployLister.Deployments("default").List(labels.Everything())
	//遍历deployment list
	for index, deploy := range deployment {
		fmt.Printf("index: %d, name: %s\n", index, deploy.GetName())
	}
	<-stopper
}
func onAdd(obj interface{}) {
	deploy := obj.(*v1.Deployment)
	klog.Info("add deployment: ", deploy.GetName())
}
func onUpdate(old, new interface{}) {
	oldDeploy := old.(*v1.Deployment)
	newDeploy := new.(*v1.Deployment)
	klog.Info("update deployment: ", oldDeploy.Status.Replicas, "->", newDeploy.Status.Replicas)
}
func onDelete(obj interface{}) {
	deploy := obj.(*v1.Deployment)
	klog.Info("delete deployment: ", deploy.GetName())
}
