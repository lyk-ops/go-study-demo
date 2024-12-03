package apis

import (
	"gin-client-go/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetPods(c *gin.Context) {
	pods, err := service.GetPods()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())

	}
	c.JSON(http.StatusOK, pods)

}
func ExecContainer(c *gin.Context) {
	namespaceName := c.Param("namespace")
	podName := c.Param("podName")
	containerName := c.Param("containerName")
	method := c.DefaultQuery("action", "sh")
	err := service.WebSSH(namespaceName, podName, containerName, method, c.Writer, c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
}
