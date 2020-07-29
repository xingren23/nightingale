package routes

import (
	"github.com/didi/nightingale/src/modules/monapi/mcache"
	"github.com/gin-gonic/gin"
)

func monitorItemGet(c *gin.Context) {
	renderData(c, mcache.MonitorItemCache.GetAll(), nil)
}
