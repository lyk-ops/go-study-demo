package main

import (
	"fmt"
	"gin-client-go/pkg/config"
	"gin-client-go/router"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

func main() {
	//var kubeConfig *string
	//ctx := context.Background()
	//if home := homedir.HomeDir(); home != "" {
	//	kubeConfig = flag.String("kubeconfig", ".kube/config", "absolute path to the kubeconfig file")
	//} else {
	//	kubeConfig = flag.String("kubeConfig", "", "absolute path to the kubeconfig file")
	//}
	//flag.Parse()
	//config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	//if err != nil {
	//	klog.Fatal(err)
	//	return
	//}
	//clientSet, err := kubernetes.NewForConfig(config)
	//if err != nil {
	//	klog.Fatal(err)
	//	return
	//}
	//namespaceList, err := clientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	//if err != nil {
	//	klog.Fatal(err)
	//	return
	//}
	//for _, item := range namespaceList.Items {
	//	fmt.Printf("namespace:%v\n", item.Name)
	//}
	engine := gin.Default()
	gin.SetMode(gin.DebugMode)
	router.InitRouter(engine)
	err := engine.Run(fmt.Sprintf("%s:%d", config.GetString(config.ServerHost), config.GetInt(config.ServerPort)))
	if err != nil {
		klog.Fatal(err)
		return
	}

}
