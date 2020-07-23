package routes

import (
	"github.com/didi/nightingale/src/modules/monapi/cmdb"
	"github.com/gin-gonic/gin"
)

func treeGet(c *gin.Context) {
	nodes, err := cmdb.GetCmdb().NodeGets("")
	renderData(c, nodes, err)
}

func treeSearchGet(c *gin.Context) {
	query := queryStr(c, "query", "")
	nodes, err := cmdb.GetCmdb().TreeSearchByPath(query)
	renderData(c, nodes, err)
}
