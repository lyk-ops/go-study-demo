package apis

import (
	"gin-client-go/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetNamespace(c *gin.Context) {
	namespace, err := service.GetNamespace()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
	}
	c.JSON(http.StatusOK, namespace)
}
