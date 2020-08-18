package routes

import (
	"strings"

	"github.com/didi/nightingale/src/modules/monapi/cmdb"
	"github.com/didi/nightingale/src/modules/monapi/scache"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"
)

type resourceForm struct {
	Nid   int64  `json:"nid"`
	Limit int    `json:"limit"`
	Host  string `json:"host"`
}

type data struct {
	Ip   string `json:"ip"`
	Type string `json:"type"`
}

func resourcePost(c *gin.Context) {
	var f resourceForm
	errors.Dangerous(c.ShouldBind(&f))

	leafNids, err := scache.GetLeafNids(f.Nid, []int64{}, []string{}, []string{})
	errors.Dangerous(err)

	endpoints, err := cmdb.GetCmdb().EndpointUnderLeafs(leafNids)
	errors.Dangerous(err)

	list := make([]data, 0)
	for idx, e := range endpoints {
		if !strings.Contains(e.Ident, f.Host) {
			continue
		}
		list = append(list, data{Ip: e.Ident, Type: e.Alias})
		if idx >= f.Limit {
			break
		}
	}

	renderData(c, list, nil)
}
