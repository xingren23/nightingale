package routes

import (
	"errors"
	"github.com/didi/nightingale/src/modules/monapi/cmdb"
	"github.com/didi/nightingale/src/modules/monapi/cmdb/meicai"
	"github.com/gin-gonic/gin"
)

func syncOps(c *gin.Context) {
	m, isMeicai := cmdb.GetCmdb().(*meicai.Meicai)
	if !isMeicai {
		renderData(c, "ok", errors.New("'meicai.Meicai' does not implement 'ICmdb'"))
	}
	renderData(c, "ok", m.SyncOps())
}
