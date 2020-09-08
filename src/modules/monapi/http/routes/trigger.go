package routes

import (
	"github.com/didi/nightingale/src/modules/monapi/cmdb/meicai"
	"github.com/gin-gonic/gin"
)

func syncOps(c *gin.Context) {
	renderData(c, "ok", meicai.M.SyncOps())
}
