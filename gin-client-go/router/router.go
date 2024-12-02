package router

import (
	"gin-client-go/apis"
	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	r.GET("/ping", apis.Ping)
	r.GET("/namespace", apis.GetNamespace)
	r.GET("/pods", apis.GetPods)
}
