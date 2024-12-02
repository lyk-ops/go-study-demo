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
