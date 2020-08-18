package routes

import (
	"github.com/didi/nightingale/src/modules/monapi/cmdb"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"
)

type resourceForm struct {
	Nid   int64  `json:"nid"`
	Limit int    `json:"limit"`
	Query string `json:"query"`
}

type data struct {
	Ip    string `json:"ip"`
	Alias string `json:"alias"`
}

func resourcePost(c *gin.Context) {
	var f resourceForm
	errors.Dangerous(c.ShouldBind(&f))

	curNode, err := cmdb.GetCmdb().NodeGet("id", f.Nid)
	errors.Dangerous(err)

	leafNids, err := cmdb.GetCmdb().LeafIds(curNode)
	errors.Dangerous(err)

	endpoints, _, err := cmdb.GetCmdb().EndpointUnderNodeGets(leafNids, f.Query, "", "", f.Limit, offset(c, f.Limit,
		999))
	errors.Dangerous(err)

	list := make([]data, 0)
	for _, e := range endpoints {
		list = append(list, data{Ip: e.Ident, Alias: e.Alias})
	}

	renderData(c, list, nil)
}
