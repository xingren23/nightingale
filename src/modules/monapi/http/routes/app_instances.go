package routes

import (
	"github.com/didi/nightingale/src/modules/monapi/cmdb"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"
)

func appInstanceGets(c *gin.Context) {
	limit := queryInt(c, "limit", 20)
	query := queryStr(c, "query", "")
	batch := queryStr(c, "batch", "")
	field := queryStr(c, "field", "ident")

	if !(field == "ident" || field == "app") {
		errors.Bomb("field invalid")
	}

	list, total, err := cmdb.GetCmdb().AppInstanceGets(query, batch, field, limit, offset(c, limit, 999))
	errors.Dangerous(err)

	renderData(c, gin.H{
		"list":  list,
		"total": total,
	}, nil)
}
